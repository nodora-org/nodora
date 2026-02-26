package nir

import (
	"encoding/json"
	"fmt"
	"strings"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/registry"
)

type Metadata map[string]any

type Program struct {
	Metadata Metadata          `json:"meta"`
	Signals  map[string]Signal `json:"signals"`
	Rules    map[string]Rule   `json:"rules"`
}

func (p *Program) GetSignal(name string) (*Signal, bool) {
	signal, exists := p.Signals[name]
	if !exists {
		return nil, false
	}
	return &signal, true
}

func (p *Program) GetRule(name string) (*Rule, bool) {
	rule, exists := p.Rules[name]
	if !exists {
		return nil, false
	}
	return &rule, true
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

type EvaluationContext struct {
	Slots     []core.Value
	Emissions []EmittedSignal
}

type Expr interface {
	Evaluate(ctx *EvaluationContext) (core.Value, error)
}

type RawExpr struct {
	Expr
}

type ImmExpr struct {
	Value core.Value `json:"imm"`
}

func (i *ImmExpr) Evaluate(ctx *EvaluationContext) (core.Value, error) {
	return i.Value, nil
}

type ArrExpr struct {
	Value []RawExpr `json:"arr"`
}

func (i *ArrExpr) Evaluate(ctx *EvaluationContext) (core.Value, error) {
	var vals []core.Value
	for _, v := range i.Value {
		ev, err := v.Evaluate(ctx)
		if err != nil {
			return core.U(), err
		}
		vals = append(vals, ev)
	}
	return core.V(vals), nil
}

type ObjExpr struct {
	Value map[string]RawExpr `json:"obj"`
}

func (i *ObjExpr) Evaluate(ctx *EvaluationContext) (core.Value, error) {
	obj := make(core.ValueMap, len(i.Value))
	for k, v := range i.Value {
		ev, err := v.Evaluate(ctx)
		if err != nil {
			return core.U(), err
		}
		if ev.Undefined {
			return core.U(), nil
		}
		obj[k] = ev
	}
	return core.V(obj), nil
}

type SymExpr struct {
	Index int `json:"sym"`
}

func (s *SymExpr) Evaluate(ctx *EvaluationContext) (core.Value, error) {
	if s.Index < 0 || s.Index >= len(ctx.Slots) {
		return core.U(), fmt.Errorf("symbol index %d out of bounds", s.Index)
	}
	return ctx.Slots[s.Index], nil
}

type IdxExpr struct {
	Index RawExpr `json:"idx"`
	From  RawExpr `json:"from"`
}

func (i *IdxExpr) Evaluate(ctx *EvaluationContext) (core.Value, error) {
	from, err := i.From.Evaluate(ctx)
	if err != nil {
		return core.U(), err
	}
	idx, err := i.Index.Evaluate(ctx)
	if err != nil {
		return core.U(), err
	}
	return evaluateAccess(from, idx)
}

type SelExpr struct {
	Path  string    `json:"path"`
	From  RawExpr   `json:"from"`
	Exprs []RawExpr `json:"$,omitempty"`
}

func (s *SelExpr) Evaluate(ctx *EvaluationContext) (core.Value, error) {
	keys := strings.Split(s.Path, ".")
	curr, err := s.From.Evaluate(ctx)
	if err != nil {
		return core.U(), err
	}

	exprIdx := 0
	for _, key := range keys {
		if curr.Undefined {
			return core.U(), nil // propagate undefined
		}
		if key == "$" {
			val, err := s.Exprs[exprIdx].Evaluate(ctx)
			if err != nil {
				return core.U(), err
			}
			curr, err = evaluateAccess(curr, val)
			if err != nil {
				return curr, err
			}
			exprIdx++
		} else {
			curr, err = evaluateAccess(curr, core.V(key))
			if err != nil {
				return curr, err
			}
		}
	}

	return curr, nil
}

func evaluateAccess(target core.Value, val core.Value) (core.Value, error) {
	if target.Undefined || val.Undefined {
		return core.U(), nil
	}
	switch t := target.Raw.(type) {
	case []core.Value:
		return evaluateArrayAccess(t, val)
	case core.ValueMap:
		return evaluateObjectAccess(t, val)
	}

	return core.U(), fmt.Errorf("cannot access into %v with value of type %v", target.Type(), val.Type())
}

func evaluateArrayAccess(from []core.Value, indexVal core.Value) (core.Value, error) {
	if indexVal.Undefined {
		return core.U(), nil
	}
	if !core.IsInt(indexVal.Raw) {
		return core.U(), fmt.Errorf("array index must be int, got %s", indexVal.Type())
	}
	idx, _ := core.ToInt(indexVal)
	if idx < 0 || idx >= len(from) {
		return core.U(), fmt.Errorf("index %d out of bounds (length %d)", idx, len(from))
	}
	return from[idx], nil
}

func evaluateObjectAccess(from core.ValueMap, keyVal core.Value) (core.Value, error) {
	if keyVal.Undefined {
		return core.U(), nil
	}
	key, ok := keyVal.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf(
			"field key must be string, got %s",
			keyVal.Type(),
		)
	}
	val, exists := from[key]
	if !exists {
		return core.U(), nil
	}
	return val, nil
}

type SignalExpr struct {
	Name string    `json:"signal"`
	Args []RawExpr `json:"args"`
	When *RawExpr  `json:"when,omitempty"`
}

func (s *SignalExpr) Evaluate(ctx *EvaluationContext) (core.Value, error) {
	if s.When != nil {
		val, err := (*s.When).Evaluate(ctx)
		if err != nil {
			return core.U(), err
		}
		if condBool, ok := val.Raw.(bool); !ok || !condBool {
			return core.U(), nil // do not emit signal if condition is false
		}
	}
	args := make([]any, len(s.Args))
	for i, arg := range s.Args {
		val, err := arg.Evaluate(ctx)
		if err != nil {
			return core.U(), err
		}
		args[i] = val.ToRaw()
	}

	ctx.Emissions = append(ctx.Emissions, EmittedSignal{
		Name: s.Name,
		Args: args,
	})
	return core.U(), nil
}

type EmittedSignal struct {
	Name string `json:"name"`
	Args []any  `json:"args"`
}

type CallFunc struct {
	Namespace string `json:"ns,omitempty"`
	Name      string `json:"name"`
}

type CallExpr struct {
	Func CallFunc  `json:"call"`
	Args []RawExpr `json:"args"`
}

func (c *CallExpr) Evaluate(ctx *EvaluationContext) (core.Value, error) {
	args := make([]core.Value, len(c.Args))
	for i, arg := range c.Args {
		val, err := arg.Evaluate(ctx)
		if err != nil {
			return core.U(), err
		}
		args[i] = val
	}
	fn, ok := registry.Global().Get(c.Func.Namespace, c.Func.Name)
	if !ok {
		return core.U(), fmt.Errorf("undefined function '%s'", fn.FullPath())
	}

	reqArgCount := fn.RequiredArgCount()
	if len(args) < reqArgCount {
		return core.U(), fmt.Errorf(
			"function '%s' expects %d argument(s), got %d",
			fn.FullPath(),
			reqArgCount,
			len(args),
		)
	}

	val, err := fn.Fn(args)
	if err != nil {
		return core.U(), fmt.Errorf("%v: %v", fn.FullPath(), err)
	}

	return val, nil
}

type LambdaParams struct {
	SymIndex int `json:"sym"`
}

type LambdaExpr struct {
	Params []LambdaParams `json:"params"`
	Ops    []Op           `json:"ops"`
}

func (le *LambdaExpr) Evaluate(ctx *EvaluationContext) (core.Value, error) {
	lambda := core.Lambda{
		Fn: func(args []core.Value) (core.Value, error) {
			return le.Invoke(ctx, args)
		},
	}
	return core.V(&lambda), nil
}

func (le *LambdaExpr) Invoke(ctx *EvaluationContext, args []core.Value) (core.Value, error) {
	for i, p := range le.Params {
		slotIdx := p.SymIndex
		ctx.Slots[slotIdx] = args[i]
	}

	var lastOut *int
	for i := range le.Ops {
		op := &le.Ops[i]
		if err := op.Execute(ctx); err != nil {
			return core.U(), err
		}
		if op.Out != nil {
			lastOut = op.Out
		}
	}

	if lastOut != nil {
		return ctx.Slots[*lastOut], nil
	}
	return core.U(), nil
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
		if condVal.Undefined {
			ctx.Slots[*op.Out] = core.U()
			return nil
		}
		condBool, ok := condVal.Raw.(bool)
		if !ok {
			return fmt.Errorf("select condition must be a boolean, got %v", condVal.Type())
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
	if val.Undefined {
		ctx.Slots[*op.Out] = core.U()
		return nil
	}
	switch op.Kind {
	case OpNot:
		b, ok := val.Raw.(bool)
		if !ok {
			return fmt.Errorf("invalid unary operand: %v", val.Type())
		}
		ctx.Slots[*op.Out] = core.V(!b)
		return nil
	default:
		return fmt.Errorf("invalid unary operation: %s %v", op.Kind, val.Raw)
	}
}

func performFloatOp(l, r float64, op *Op, ctx *EvaluationContext) error {
	switch op.Kind {
	case OpAdd:
		ctx.Slots[*op.Out] = core.V(l + r)
	case OpSub:
		ctx.Slots[*op.Out] = core.V(l - r)
	case OpMul:
		ctx.Slots[*op.Out] = core.V(l * r)
	case OpDiv:
		if r == 0 {
			ctx.Slots[*op.Out] = core.U() // div by zero is undefined
			return nil
		}
		ctx.Slots[*op.Out] = core.V(l / r)
	case OpMod:
		if r == 0 {
			ctx.Slots[*op.Out] = core.U() // mod (div) by zero is undefined
			return nil
		}
		ctx.Slots[*op.Out] = core.V(float64(int64(l) % int64(r)))
	case OpLt:
		ctx.Slots[*op.Out] = core.V(l < r)
	case OpGt:
		ctx.Slots[*op.Out] = core.V(l > r)
	case OpLte:
		ctx.Slots[*op.Out] = core.V(l <= r)
	case OpGte:
		ctx.Slots[*op.Out] = core.V(l >= r)
	default:
		return fmt.Errorf("invalid operation: %v %s %v", l, op.Kind, r)
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
		ctx.Slots[*op.Out] = core.U()
		return nil
	}

	// handle operations that work on any type
	switch op.Kind {
	case OpEq:
		ctx.Slots[*op.Out] = core.V(core.SafeEquals(lval, rval))
		return nil
	case OpNeq:
		ctx.Slots[*op.Out] = core.V(!core.SafeEquals(lval, rval))
		return nil
	case OpIn:
		arr, ok := rval.Raw.([]core.Value)
		if !ok {
			return fmt.Errorf("right operand of '%v' must be an array, got %v", OpIn, rval.Type())
		}
		ctx.Slots[*op.Out] = core.V(contains(arr, lval))
		return nil
	}

	// handle boolean operations
	if l, ok := lval.Raw.(bool); ok {
		r, ok := rval.Raw.(bool)
		if !ok {
			return fmt.Errorf(
				"operator '%s' requires both operands to be of type bool, got %v and %v",
				op.Kind,
				lval.Type(),
				rval.Type(),
			)
		}
		switch op.Kind {
		case OpAnd:
			ctx.Slots[*op.Out] = core.V(l && r)
		case OpOr:
			ctx.Slots[*op.Out] = core.V(l || r)
		default:
			return fmt.Errorf("operator '%s' is not supported for type bool", op.Kind)
		}
		return nil
	}

	// handle string operations
	if l, ok := lval.Raw.(string); ok {
		switch op.Kind {
		case OpAdd:
			r, ok := rval.Raw.(string)
			if !ok {
				ctx.Slots[*op.Out] = core.U()
				return fmt.Errorf(
					"operator '%s' requires both operands to be of type string, got %v and %v",
					op.Kind,
					lval.Type(),
					rval.Type(),
				)
			}
			ctx.Slots[*op.Out] = core.V(l + r)
			return nil
		default:
			return fmt.Errorf("operator '%s' is not supported for type string", op.Kind)
		}
	}

	l, lok := core.ToFloat64(lval)
	if !lok {
		return fmt.Errorf("operator '%s' cannot convert left operand to float64: %v", op.Kind, lval.Type())
	}
	r, rok := core.ToFloat64(rval)
	if !rok {
		return fmt.Errorf("operator '%s' cannot convert right operand to float64: %v", op.Kind, rval.Type())
	}
	return performFloatOp(l, r, op, ctx)
}

func (w RawExpr) MarshalJSON() ([]byte, error) {
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
	case probe["obj"] != nil:
		var e ObjExpr
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
	case probe["call"] != nil:
		var e CallExpr
		if err := json.Unmarshal(data, &e); err != nil {
			return err
		}
		w.Expr = &e
	case probe["ops"] != nil:
		var e LambdaExpr
		if err := json.Unmarshal(data, &e); err != nil {
			return err
		}
		w.Expr = &e
	default:
		return fmt.Errorf("unknown shape: %s", string(data))
	}
	return nil
}
