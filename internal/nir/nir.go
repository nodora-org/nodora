package nir

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Program struct {
	Signals map[string]Signal `json:"signals"`
	Rules   map[string]Rule   `json:"rules"`
}

type Signal struct {
	Params []Param `json:"params"`
}

type Param struct {
	Name string `json:"name"`
}

type Value any

type SymbolMap map[string]int

type Output struct {
	Sym int `json:"sym"`
}

type Rule struct {
	Outputs  map[string]Output `json:"outputs"`
	Symslots int               `json:"symslots"`
	Symbols  SymbolMap         `json:"symbols"`
	Ops      []Op              `json:"ops"`
}

type SignalListener interface {
	Invoke(args []Value) error
}

type SignalListenerFunc func(args []Value) error

func (f SignalListenerFunc) Invoke(args []Value) error {
	return f(args)
}

type EvaluationContext struct {
	Input     map[string]any
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

type InputExpr struct {
	Value string `json:"input"`
}

func (s *InputExpr) Evaluate(ctx *EvaluationContext) (Value, error) {
	keys := strings.Split(s.Value, ".")
	current := any(ctx.Input)
	for _, key := range keys {
		m, ok := current.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("cannot access field '%s': parent is not an object", key)
		}
		val, exists := m[key]
		if !exists {
			return nil, fmt.Errorf("field '%s' does not exist in path '%s'", key, s.Value)
		}
		current = val
	}
	return current, nil
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
		for _, listener := range listeners {
			listener.Invoke(args)
		}
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
		if len(o.Args) != 1 {
			return fmt.Errorf("invalid operation arguments for '%s'", o.Kind)
		}
		if o.Out == nil {
			return fmt.Errorf("output index required for '%s'", o.Kind)
		}
		ctx.Slots[*o.Out], _ = o.Args[0].Expr.Evaluate(ctx)
		return nil
	case OpAdd, OpSub, OpMul, OpDiv, OpMod, OpIn, OpAnd, OpOr, OpLt, OpGt, OpLte, OpGte, OpEq, OpNeq:
		if len(o.Args) != 2 {
			return fmt.Errorf("invalid operation arguments for '%s'", o.Kind)
		}
		if o.Out == nil {
			return fmt.Errorf("output index required for '%s'", o.Kind)
		}
		if *o.Out < 0 || *o.Out >= len(ctx.Slots) {
			return fmt.Errorf("output index %d is out of bounds", *o.Out)
		}
		return executeBinaryOp(o, ctx)
	case OpNot:
		if len(o.Args) != 1 {
			return fmt.Errorf("invalid operation arguments for '%s'", o.Kind)
		}
		if o.Out == nil {
			return fmt.Errorf("output index required for '%s'", o.Kind)
		}
		if *o.Out < 0 || *o.Out >= len(ctx.Slots) {
			return fmt.Errorf("output index %d is out of bounds", *o.Out)
		}
		return executeUnaryOp(o, ctx)
	case OpSelect:
		if len(o.Args) != 3 {
			return fmt.Errorf("invalid operation arguments for '%s', expected 3 args (condition, true_val, false_val)", o.Kind)
		}
		if o.Out == nil {
			return fmt.Errorf("output index required for '%s'", o.Kind)
		}
		if *o.Out < 0 || *o.Out >= len(ctx.Slots) {
			return fmt.Errorf("output index %d is out of bounds", *o.Out)
		}
		condVal, err := o.Args[0].Evaluate(ctx)
		if err != nil {
			return err
		}
		condBool, ok := condVal.(bool)
		if !ok {
			return fmt.Errorf("select condition must be boolean, got %T", condVal)
		}
		if condBool {
			ctx.Slots[*o.Out], _ = o.Args[1].Evaluate(ctx)
		} else {
			ctx.Slots[*o.Out], _ = o.Args[2].Evaluate(ctx)
		}
		return nil
	case OpEmit:
		if len(o.Args) != 1 {
			return fmt.Errorf("invalid operation arguments for '%s'", o.Kind)
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
	return fmt.Errorf("unsupported operation: %s", o.Kind)
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
		if r, ok := rval.(bool); ok {
			switch op.Kind {
			case OpAnd:
				ctx.Slots[*op.Out] = l && r
			case OpOr:
				ctx.Slots[*op.Out] = l || r
			default:
				return fmt.Errorf("invalid boolean operation: %v %s %v", lval, op.Kind, rval)
			}
			return nil
		}
		return fmt.Errorf("cannot perform boolean operation %s with non-boolean operand: %T", op.Kind, rval)
	}

	// handle string operations
	if l, ok := lval.(string); ok {
		if r, ok := rval.(string); ok {
			switch op.Kind {
			case OpAdd:
				ctx.Slots[*op.Out] = l + r
				return nil
			default:
				return fmt.Errorf("invalid string operation: %v %s %v", lval, op.Kind, rval)
			}
		}
		return fmt.Errorf("cannot perform string operation %s with non-string operand: %T", op.Kind, rval)
	}

	if isFloat(lval) || isFloat(rval) {
		l, lok := toFloat64(lval)
		r, rok := toFloat64(rval)
		if !lok || !rok {
			return fmt.Errorf("cannot convert operands to float64: %T %s %T", lval, op.Kind, rval)
		}
		return performNumericOp(l, r, op, ctx)
	}

	l, lok := toInt64(lval)
	r, rok := toInt64(rval)
	if !lok || !rok {
		return fmt.Errorf("cannot convert operands to int64: %T %s %T", lval, op.Kind, rval)
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
	case probe["input"] != nil:
		var e InputExpr
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
