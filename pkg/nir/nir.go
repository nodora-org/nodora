package nir

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Metadata map[string]any

type Program struct {
	Metadata Metadata          `json:"meta"`
	Signals  map[string]Signal `json:"signals"`
	Rules    map[string]Rule   `json:"rules"`
}

type Signal struct {
	Params []Param `json:"params"`
}

type Param struct {
	Name string `json:"name"`
}

type Output struct {
	Sym int `json:"sym"`
}

type Rule struct {
	Outputs  map[string]Output `json:"outputs"`
	Symslots int               `json:"symslots"`
	Ops      []Op              `json:"ops"`
}

type SignalListener interface {
	Invoke(args []any) error
}

type SignalListenerFunc func(args []any) error

func (f SignalListenerFunc) Invoke(args []any) error {
	defer func() { recover() }() // silent recovery
	return f(args)
}

type EvaluationContext struct {
	Slots     []Value
	Emissions []EmittedSignal
	Errors    []error
	Listeners map[string][]SignalListener
	OpIndex   *int
}

func (ctx *EvaluationContext) addError(err error) {
	ctx.Errors = append(ctx.Errors, fmt.Errorf("%d: %v", *ctx.OpIndex, err))
}

type Expr interface {
	Evaluate(ctx *EvaluationContext) (Value, error)
}

type RawExpr struct {
	Expr
}

type ImmExpr struct {
	Value Value `json:"imm"`
}

func (i *ImmExpr) Evaluate(ctx *EvaluationContext) (Value, error) {
	return i.Value, nil
}

type ArrExpr struct {
	Value []RawExpr `json:"arr"`
}

func (i *ArrExpr) Evaluate(ctx *EvaluationContext) (Value, error) {
	var vals []Value
	for _, v := range i.Value {
		ev, err := v.Evaluate(ctx)
		if err != nil {
			return U(), err
		}
		vals = append(vals, ev)
	}
	return V(vals), nil
}

type SymExpr struct {
	Index int `json:"sym"`
}

func (s *SymExpr) Evaluate(ctx *EvaluationContext) (Value, error) {
	if s.Index < 0 || s.Index >= len(ctx.Slots) {
		return U(), fmt.Errorf("symbol index %d out of bounds", s.Index)
	}
	return ctx.Slots[s.Index], nil
}

type IdxExpr struct {
	Index RawExpr `json:"idx"`
	From  RawExpr `json:"from"`
}

func (i *IdxExpr) Evaluate(ctx *EvaluationContext) (Value, error) {
	from, err := i.From.Evaluate(ctx)
	if err != nil {
		return U(), err
	}
	idx, err := i.Index.Evaluate(ctx)
	if err != nil {
		return U(), err
	}
	return evaluateAccess(ctx, from, idx), nil
}

type SelExpr struct {
	Path  string    `json:"path"`
	From  RawExpr   `json:"from"`
	Exprs []RawExpr `json:"$,omitempty"`
}

func (s *SelExpr) Evaluate(ctx *EvaluationContext) (Value, error) {
	keys := strings.Split(s.Path, ".")
	current, err := s.From.Evaluate(ctx)
	if err != nil {
		return U(), err
	}

	exprIdx := 0
	for _, key := range keys {
		if current.Undefined {
			return U(), nil // propagate undefined
		}
		if key == "$" {
			val, err := s.Exprs[exprIdx].Evaluate(ctx)
			if err != nil {
				return U(), err
			}
			current = evaluateAccess(ctx, current, val)
			exprIdx++
		} else {
			current = evaluateAccess(ctx, current, V(key))
		}
	}

	return current, nil
}

func evaluateAccess(ctx *EvaluationContext, target Value, val Value) Value {
	if target.Undefined || val.Undefined {
		return U()
	}
	switch t := target.Raw.(type) {
	case []Value:
		return evaluateArrayAccess(ctx, t, val)
	case ValueMap:
		return evaluateObjectAccess(ctx, t, val)
	}
	ctx.addError(fmt.Errorf("cannot access into %v with value of type %v", target.Type(), val.Type()))
	return U()
}

func evaluateArrayAccess(ctx *EvaluationContext, from []Value, indexVal Value) Value {
	if indexVal.Undefined {
		return U()
	}
	idx, ok := toInt(indexVal)
	if !ok {
		ctx.addError(fmt.Errorf(
			"array index must be int, got %s",
			indexVal.Type(),
		))
		return U()
	}
	if idx < 0 || idx >= len(from) {
		ctx.addError(fmt.Errorf(
			"index %d out of bounds (length %d)",
			idx,
			len(from),
		))
		return U()
	}
	return from[idx]
}

func evaluateObjectAccess(ctx *EvaluationContext, from ValueMap, keyVal Value) Value {
	if keyVal.Undefined {
		return U()
	}
	key, ok := keyVal.Raw.(string)
	if !ok {
		ctx.addError(fmt.Errorf(
			"field key must be string, got %s",
			keyVal.Type(),
		))
		return U()
	}

	val, exists := from[key]
	if !exists {
		return U()
	}

	return val
}

type SignalExpr struct {
	Name string    `json:"signal"`
	Args []RawExpr `json:"args"`
	When *RawExpr  `json:"when,omitempty"`
}

func (s *SignalExpr) Evaluate(ctx *EvaluationContext) (Value, error) {
	if s.When != nil {
		val, err := (*s.When).Evaluate(ctx)
		if err != nil {
			return U(), err
		}
		if condBool, ok := val.Raw.(bool); !ok || !condBool {
			return U(), nil // do not emit signal if condition is false
		}
	}
	args := make([]any, len(s.Args))
	for i, arg := range s.Args {
		val, err := arg.Evaluate(ctx)
		if err != nil {
			return U(), err
		}
		args[i] = val.ToRaw()
	}

	emittedSignal := EmittedSignal{
		Name: s.Name,
		Args: args,
	}

	ctx.Emissions = append(ctx.Emissions, emittedSignal)
	if listeners, exists := ctx.Listeners[s.Name]; exists {
		s.invokeListeners(listeners, args)
	}

	return U(), nil
}

type EmittedSignal struct {
	Name string `json:"name"`
	Args []any  `json:"args"`
}

type OpKind string

const (
	OpCopy   OpKind = "cp"
	OpAdd    OpKind = "add"
	OpSub    OpKind = "sub"
	OpMul    OpKind = "mul"
	OpDiv    OpKind = "div"
	OpMod    OpKind = "mod"
	OpAnd    OpKind = "and"
	OpOr     OpKind = "or"
	OpGt     OpKind = "gt"
	OpGte    OpKind = "gte"
	OpLt     OpKind = "lt"
	OpLte    OpKind = "lte"
	OpEq     OpKind = "eq"
	OpNeq    OpKind = "neq"
	OpIn     OpKind = "in"
	OpNot    OpKind = "not"
	OpSelect OpKind = "select"
	OpEmit   OpKind = "emit"
)

type Op struct {
	Kind OpKind    `json:"op"`
	Args []RawExpr `json:"args"`
	Out  *int      `json:"out,omitempty"`
}

func (op *Op) Execute(ctx *EvaluationContext) error {
	switch op.Kind {
	case OpCopy:
		if err := validateOp(op, ctx, 1); err != nil {
			return err
		}
		val, err := op.Args[0].Evaluate(ctx)
		if err != nil {
			return err
		}
		ctx.Slots[*op.Out] = val
		return nil
	case OpAdd, OpSub, OpMul, OpDiv, OpMod, OpIn, OpAnd, OpOr, OpLt, OpGt, OpLte, OpGte, OpEq, OpNeq:
		if err := validateOp(op, ctx, 2); err != nil {
			return err
		}
		return executeBinaryOp(op, ctx)
	case OpNot:
		if err := validateOp(op, ctx, 1); err != nil {
			return err
		}
		return executeUnaryOp(op, ctx)
	case OpSelect:
		if err := validateOp(op, ctx, 3); err != nil {
			return err
		}
		condVal, err := op.Args[0].Evaluate(ctx)
		if err != nil {
			return err
		}
		condBool, ok := condVal.Raw.(bool)
		if !ok {
			ctx.Slots[*op.Out] = U()
			ctx.addError(fmt.Errorf("select condition must be a boolean, got %v", condVal.Type()))
			return nil
		}
		idx := 1
		if !condBool {
			idx = 2
		}
		val, err := op.Args[idx].Evaluate(ctx)
		if err != nil {
			return err
		}
		ctx.Slots[*op.Out] = val
		return nil
	case OpEmit:
		if err := validateArgs(op, 1); err != nil {
			return err
		}
		_, err := op.Args[0].Evaluate(ctx)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("unknown operation '%s'", op.Kind)
}

func executeUnaryOp(op *Op, ctx *EvaluationContext) error {
	val, err := op.Args[0].Evaluate(ctx)
	if err != nil {
		return err
	}

	switch op.Kind {
	case OpNot:
		b, ok := val.Raw.(bool)
		if !ok {
			ctx.Slots[*op.Out] = U()
			ctx.addError(fmt.Errorf("invalid unary operand: %v", val.Type()))
			return nil
		}
		ctx.Slots[*op.Out] = V(!b)
		return nil
	default:
		return fmt.Errorf("invalid unary operation: %s %v", op.Kind, val.Raw)
	}
}

func performNumericOp[T Numeric](l T, r T, op *Op, ctx *EvaluationContext) error {
	switch op.Kind {
	case OpAdd:
		ctx.Slots[*op.Out] = V(l + r)
	case OpSub:
		ctx.Slots[*op.Out] = V(l - r)
	case OpMul:
		ctx.Slots[*op.Out] = V(l * r)
	case OpDiv:
		if r == 0 {
			ctx.Slots[*op.Out] = U() // div by zero is undefined
			return nil
		}
		ctx.Slots[*op.Out] = V(l / r)
	case OpMod:
		if r == 0 {
			ctx.Slots[*op.Out] = U() // mod (div) by zero is undefined
			return nil
		}
		ctx.Slots[*op.Out] = V(T(int64(l) % int64(r)))
	case OpLt:
		ctx.Slots[*op.Out] = V(l < r)
	case OpGt:
		ctx.Slots[*op.Out] = V(l > r)
	case OpLte:
		ctx.Slots[*op.Out] = V(l <= r)
	case OpGte:
		ctx.Slots[*op.Out] = V(l >= r)
	default:
		ctx.Slots[*op.Out] = U()
		ctx.addError(fmt.Errorf("invalid numeric operation: %v %s %v", l, op.Kind, r))
		return nil
	}
	return nil
}

func executeBinaryOp(op *Op, ctx *EvaluationContext) error {
	lval, err := op.Args[0].Evaluate(ctx)
	if err != nil {
		return err
	}
	rval, err := op.Args[1].Evaluate(ctx)
	if err != nil {
		return err
	}

	// if either side is undefined, result is undefined
	if lval.Undefined || rval.Undefined {
		ctx.Slots[*op.Out] = U()
		return nil
	}

	// handle operations that work on any type
	switch op.Kind {
	case OpEq:
		ctx.Slots[*op.Out] = V(lval.Raw == rval.Raw)
		return nil
	case OpNeq:
		ctx.Slots[*op.Out] = V(lval.Raw != rval.Raw)
		return nil
	case OpIn:
		arr, ok := rval.Raw.([]Value)
		if !ok {
			ctx.Slots[*op.Out] = U()
			ctx.addError(fmt.Errorf("right operand of '%v' must be an array, got %v", OpIn, rval.Type()))
			return nil
		}
		ctx.Slots[*op.Out] = V(contains(arr, lval.Raw))
		return nil
	}

	// handle boolean operations
	if l, ok := lval.Raw.(bool); ok {
		r, ok := rval.Raw.(bool)
		if !ok {
			ctx.Slots[*op.Out] = U()
			ctx.addError(fmt.Errorf(
				"operator '%s' requires both operands to be bool, got %v and %v",
				op.Kind,
				lval.Type(),
				rval.Type(),
			))
			return nil
		}
		switch op.Kind {
		case OpAnd:
			ctx.Slots[*op.Out] = V(l && r)
		case OpOr:
			ctx.Slots[*op.Out] = V(l || r)
		default:
			ctx.Slots[*op.Out] = U()
			ctx.addError(fmt.Errorf("operator '%s' is not supported for booleans", op.Kind))
			return nil
		}
		return nil
	}

	// handle string operations
	if l, ok := lval.Raw.(string); ok {
		r, ok := rval.Raw.(string)
		if !ok {
			ctx.Slots[*op.Out] = U()
			ctx.addError(fmt.Errorf(
				"operator '%s' requires both operands to be strings, got %v and %v",
				op.Kind,
				lval.Type(),
				rval.Type(),
			))
			return nil
		}
		switch op.Kind {
		case OpAdd:
			ctx.Slots[*op.Out] = V(l + r)
			return nil
		default:
			ctx.Slots[*op.Out] = U()
			ctx.addError(fmt.Errorf("operator '%s' is not supported for strings", op.Kind))
			return nil
		}
	}

	if isFloat(lval) || isFloat(rval) {
		l, lok := toFloat64(lval)
		if !lok {
			ctx.Slots[*op.Out] = U()
			ctx.addError(fmt.Errorf("operator '%s' cannot convert left operand to float64: %v", op.Kind, lval.Type()))
			return nil
		}
		r, rok := toFloat64(rval)
		if !rok {
			ctx.Slots[*op.Out] = U()
			ctx.addError(fmt.Errorf("operator '%s' cannot convert right operand to float64: %v", op.Kind, rval.Type()))
			return nil
		}
		return performNumericOp(l, r, op, ctx)
	}

	l, lok := toInt64(lval)
	if !lok {
		ctx.Slots[*op.Out] = U()
		ctx.addError(fmt.Errorf("operator '%s' cannot convert left operand to int64: %v", op.Kind, lval.Type()))
		return nil
	}
	r, rok := toInt64(rval)
	if !rok {
		ctx.Slots[*op.Out] = U()
		ctx.addError(fmt.Errorf("operator '%s' cannot convert right operand to int64: %v", op.Kind, rval.Type()))
		return nil
	}

	return performNumericOp(l, r, op, ctx)
}

func (w *RawExpr) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.Expr)
}

func (w *RawExpr) UnmarshalJSON(data []byte) error {
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}
	switch {
	case probe["sym"] != nil:
		var e SymExpr
		if err := json.Unmarshal(data, &e); err != nil {
			return err
		}
		w.Expr = &e
	case probe["imm"] != nil:
		var e ImmExpr
		if err := json.Unmarshal(data, &e); err != nil {
			return err
		}
		w.Expr = &e
	case probe["arr"] != nil:
		var e ArrExpr
		if err := json.Unmarshal(data, &e); err != nil {
			return err
		}
		w.Expr = &e
	case probe["idx"] != nil:
		var e IdxExpr
		if err := json.Unmarshal(data, &e); err != nil {
			return err
		}
		w.Expr = &e
	case probe["path"] != nil:
		var e SelExpr
		if err := json.Unmarshal(data, &e); err != nil {
			return err
		}
		w.Expr = &e
	case probe["signal"] != nil:
		var e SignalExpr
		if err := json.Unmarshal(data, &e); err != nil {
			return err
		}
		w.Expr = &e
	default:
		return fmt.Errorf("unknown shape: %s", string(data))
	}
	return nil
}
