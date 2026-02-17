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
	Invoke(args []Value) error
}

type SignalListenerFunc func(args []Value) error

func (f SignalListenerFunc) Invoke(args []Value) error {
	defer func() { recover() }() // silent recovery
	return f(args)
}

type EvaluationContext struct {
	Slots     []Value
	Emissions []EmittedSignal
	Listeners map[string][]SignalListener
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
			return nil, err
		}
		vals = append(vals, ev)
	}
	return vals, nil
}

type SymExpr struct {
	Index int `json:"sym"`
}

func (s *SymExpr) Evaluate(ctx *EvaluationContext) (Value, error) {
	if s.Index < 0 || s.Index >= len(ctx.Slots) {
		return nil, fmt.Errorf("symbol index %d out of bounds", s.Index)
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
		return nil, err
	}
	idx, err := i.Index.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	return evaluateAccess(from, idx)
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
		return nil, err
	}

	exprIdx := 0
	for _, key := range keys {
		var err error
		if key == "$" {
			val, err := s.Exprs[exprIdx].Evaluate(ctx)
			if err != nil {
				return nil, err
			}
			current, err = evaluateAccess(current, val)
			exprIdx++
		} else {
			current, err = evaluateAccess(current, key)
		}
		if err != nil {
			return nil, err
		}
	}

	return current, nil
}

func evaluateAccess(target Value, val Value) (Value, error) {
	switch t := target.(type) {
	case []Value:
		return evaluateArrayAccess(t, val)
	case ValueMap:
		return evaluateObjectAccess(t, val)
	}
	return nil, fmt.Errorf("cannot access into %T with value of type %T", target, val)
}

func evaluateArrayAccess(from []Value, indexVal Value) (Value, error) {
	idx, ok := toInt(indexVal)
	if !ok {
		return nil, fmt.Errorf("array index must be an int, got %T", indexVal)
	}
	if idx < 0 || idx >= len(from) {
		return nil, fmt.Errorf("index %d out of bounds (length %d)", idx, len(from))
	}
	return from[idx], nil
}

func evaluateObjectAccess(from ValueMap, keyVal Value) (Value, error) {
	key, ok := keyVal.(string)
	if !ok {
		return nil, fmt.Errorf("field key must be a string, got %T", keyVal)
	}

	val, exists := from[key]
	if !exists {
		return nil, fmt.Errorf("field '%s' does not exist on object", key)
	}

	return val, nil
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
			return nil, err
		}
		if condBool, ok := val.(bool); !ok || !condBool {
			return nil, nil // do not emit signal if condition is false
		}
	}
	args := make([]Value, len(s.Args))
	for i, arg := range s.Args {
		val, err := arg.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		args[i] = val
	}
	emittedSignal := EmittedSignal{
		Name: s.Name,
		Args: args,
	}

	if listeners, exists := ctx.Listeners[s.Name]; exists {
		s.invokeListeners(listeners, args)
	}
	return emittedSignal, nil
}

type EmittedSignal struct {
	Name string  `json:"name"`
	Args []Value `json:"args"`
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

func (o *Op) Execute(ctx *EvaluationContext) error {
	switch o.Kind {
	case OpCopy:
		if err := validateOp(o, ctx, 1); err != nil {
			return err
		}
		val, err := o.Args[0].Expr.Evaluate(ctx)
		if err != nil {
			return err
		}
		ctx.Slots[*o.Out] = val
		return nil
	case OpAdd, OpSub, OpMul, OpDiv, OpMod, OpIn, OpAnd, OpOr, OpLt, OpGt, OpLte, OpGte, OpEq, OpNeq:
		if err := validateOp(o, ctx, 2); err != nil {
			return err
		}
		return executeBinaryOp(o, ctx)
	case OpNot:
		if err := validateOp(o, ctx, 1); err != nil {
			return err
		}
		return executeUnaryOp(o, ctx)
	case OpSelect:
		if err := validateOp(o, ctx, 3); err != nil {
			return err
		}
		condVal, err := o.Args[0].Evaluate(ctx)
		if err != nil {
			return err
		}
		condBool, ok := condVal.(bool)
		if !ok {
			return fmt.Errorf("select condition must be boolean, got %T", condVal)
		}
		idx := 1
		if !condBool {
			idx = 2
		}
		val, err := o.Args[idx].Evaluate(ctx)
		if err != nil {
			return err
		}
		ctx.Slots[*o.Out] = val
		return nil
	case OpEmit:
		if err := validateArgs(o, 1); err != nil {
			return err
		}
		val, err := o.Args[0].Evaluate(ctx)
		if err != nil {
			return fmt.Errorf("emit operation evaluation error: %v", err)
		}
		if val != nil {
			ctx.Emissions = append(ctx.Emissions, val.(EmittedSignal))
		}
		return nil
	}
	return fmt.Errorf("unknown operation '%s'", o.Kind)
}

func executeUnaryOp(op *Op, ctx *EvaluationContext) error {
	val, err := op.Args[0].Evaluate(ctx)
	if err != nil {
		return err
	}

	switch op.Kind {
	case OpNot:
		if b, ok := val.(bool); ok {
			ctx.Slots[*op.Out] = !b
			return nil
		}
	default:
		return fmt.Errorf("invalid unary operation: %s %v", op.Kind, val)
	}

	return fmt.Errorf("invalid operand for 'not': %v", val)
}

func performNumericOp[T Numeric](l T, r T, op *Op, ctx *EvaluationContext) error {
	switch op.Kind {
	case OpAdd:
		ctx.Slots[*op.Out] = l + r
	case OpSub:
		ctx.Slots[*op.Out] = l - r
	case OpMul:
		ctx.Slots[*op.Out] = l * r
	case OpDiv:
		if r == 0 {
			return fmt.Errorf("division by zero")
		}
		ctx.Slots[*op.Out] = l / r
	case OpMod:
		ctx.Slots[*op.Out] = T(int64(l) % int64(r))
	case OpLt:
		ctx.Slots[*op.Out] = l < r
	case OpGt:
		ctx.Slots[*op.Out] = l > r
	case OpLte:
		ctx.Slots[*op.Out] = l <= r
	case OpGte:
		ctx.Slots[*op.Out] = l >= r
	default:
		return fmt.Errorf("invalid numeric operation: %v %s %v", l, op.Kind, r)
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

	// handle operations that work on any type
	switch op.Kind {
	case OpEq:
		ctx.Slots[*op.Out] = lval == rval
		return nil
	case OpNeq:
		ctx.Slots[*op.Out] = lval != rval
		return nil
	case OpIn:
		found, err := containsValue(rval, lval.(any))
		if err != nil {
			return fmt.Errorf("invalid right operand of '%v': %v", OpIn, err)
		}
		ctx.Slots[*op.Out] = found
		return nil
	}

	// handle boolean operations
	if l, ok := lval.(bool); ok {
		r, ok := rval.(bool)
		if !ok {
			return fmt.Errorf("operator '%s' requires both operands to be bool, got %T and %T", op.Kind, lval, rval)
		}
		switch op.Kind {
		case OpAnd:
			ctx.Slots[*op.Out] = l && r
		case OpOr:
			ctx.Slots[*op.Out] = l || r
		default:
			return fmt.Errorf("operator '%s' is not supported for booleans", op.Kind)
		}
		return nil
	}

	// handle string operations
	if l, ok := lval.(string); ok {
		r, ok := rval.(string)
		if !ok {
			return fmt.Errorf("operator '%s' requires both operands to be strings, got %T and %T", op.Kind, lval, rval)
		}
		switch op.Kind {
		case OpAdd:
			ctx.Slots[*op.Out] = l + r
			return nil
		default:
			return fmt.Errorf("operator '%s' is not supported for strings", op.Kind)
		}
	}

	if isFloat(lval) || isFloat(rval) {
		l, lok := toFloat64(lval)
		if !lok {
			return fmt.Errorf("operator '%s' cannot convert left operand to float64: %T", op.Kind, lval)
		}
		r, rok := toFloat64(rval)
		if !rok {
			return fmt.Errorf("operator '%s' cannot convert right operand to float64: %T", op.Kind, rval)
		}
		return performNumericOp(l, r, op, ctx)
	}

	l, lok := toInt64(lval)
	if !lok {
		return fmt.Errorf("operator '%s' cannot convert left operand to int64: %T", op.Kind, lval)
	}
	r, rok := toInt64(rval)
	if !rok {
		return fmt.Errorf("operator '%s' cannot convert right operand to int64: %T", op.Kind, rval)
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
