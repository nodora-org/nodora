package nir

import (
	"encoding/json"
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

type Metadata map[string]any

type Ruleset struct {
	Metadata Metadata          `json:"meta"`
	Signals  map[string]Signal `json:"signals"`
	Rules    map[string]Rule   `json:"rules"`
}

func (p *Ruleset) GetSignal(name string) (*Signal, bool) {
	signal, exists := p.Signals[name]
	if !exists {
		return nil, false
	}
	return &signal, true
}

func (p *Ruleset) GetRule(name string) (*Rule, bool) {
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
	vals := make([]core.Value, len(i.Value))
	for idx, v := range i.Value {
		ev, err := v.Evaluate(ctx)
		if err != nil {
			return core.U(), err
		}
		vals[idx] = ev
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
		if ev.IsUndefined() {
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

	// pre-split path
	keys []string
}

func (s *SelExpr) Evaluate(ctx *EvaluationContext) (core.Value, error) {
	curr, err := s.From.Evaluate(ctx)
	if err != nil {
		return core.U(), err
	}

	exprIdx := 0
	for _, key := range s.keys {
		if curr.IsUndefined() {
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
			curr, err = evaluateFieldAccess(curr, key)
			if err != nil {
				return curr, err
			}
		}
	}

	return curr, nil
}

func evaluateFieldAccess(target core.Value, key string) (core.Value, error) {
	obj, ok := target.AsObject()
	if !ok {
		return core.U(), fmt.Errorf("cannot access into %v with value of type string", target.Type())
	}
	val, exists := obj[key]
	if !exists {
		return core.U(), nil
	}
	return val, nil
}

func evaluateAccess(target core.Value, val core.Value) (core.Value, error) {
	if target.IsUndefined() || val.IsUndefined() {
		return core.U(), nil
	}
	if arr, ok := target.AsArray(); ok {
		return evaluateArrayAccess(arr, val)
	}
	if obj, ok := target.AsObject(); ok {
		return evaluateObjectAccess(obj, val)
	}

	return core.U(), fmt.Errorf("cannot access into %v with value of type %v", target.Type(), val.Type())
}

func evaluateArrayAccess(from []core.Value, indexVal core.Value) (core.Value, error) {
	if indexVal.IsUndefined() {
		return core.U(), nil
	}
	idx, ok := indexVal.AsIndex()
	if !ok {
		return core.U(), fmt.Errorf("array index must be int, got %s", indexVal.Type())
	}
	if idx < 0 || idx >= len(from) {
		return core.U(), fmt.Errorf("index %d out of bounds (length %d)", idx, len(from))
	}
	return from[idx], nil
}

func evaluateObjectAccess(from core.ValueMap, keyVal core.Value) (core.Value, error) {
	if keyVal.IsUndefined() {
		return core.U(), nil
	}
	key, ok := keyVal.AsString()
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
		if condBool, ok := val.AsBool(); !ok || !condBool {
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

func (c *CallFunc) String() string {
	if c.Namespace == "" {
		return c.Name
	}
	return c.Namespace + "::" + c.Name
}

type CallExpr struct {
	Func CallFunc  `json:"call"`
	Args []RawExpr `json:"args"`

	// resolved function
	fn *types.Func
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

	fn := c.fn

	// short-circuit to undefined if any required argument is undefined
	// skip functions that intentionally consume undefined
	if !fn.AcceptsUndefined {
		for i, arg := range args {
			if arg.IsUndefined() && i < len(fn.Args) && fn.Args[i].Required {
				return core.U(), nil
			}
		}
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

	// interned opcode
	code opCode
}

func (op *Op) Execute(ctx *EvaluationContext) error {
	switch op.code {
	case opcCopy:
		val, err := op.Args[0].Evaluate(ctx)
		if err != nil {
			return err
		}
		ctx.Slots[*op.Out] = val
		return nil
	case opcAdd, opcSub, opcMul, opcDiv, opcMod, opcIn, opcAnd, opcOr, opcLt, opcGt, opcLte, opcGte, opcEq, opcNeq:
		return executeBinaryOp(op, ctx)
	case opcNot:
		val, err := op.Args[0].Evaluate(ctx)
		if err != nil {
			return err
		}
		if val.IsUndefined() {
			ctx.Slots[*op.Out] = core.U()
			return nil
		}
		b, ok := val.AsBool()
		if !ok {
			ctx.Slots[*op.Out] = core.U()
			return nil
		}
		ctx.Slots[*op.Out] = core.Bool(!b)
		return nil
	case opcSelect:
		condVal, err := op.Args[0].Evaluate(ctx)
		if err != nil {
			return err
		}
		if condVal.IsUndefined() {
			ctx.Slots[*op.Out] = core.U()
			return nil
		}
		condBool, ok := condVal.AsBool()
		if !ok {
			ctx.Slots[*op.Out] = core.U()
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
	case opcEmit:
		_, err := op.Args[0].Evaluate(ctx)
		return err
	}
	return fmt.Errorf("unknown operation '%s'", op.Kind)
}

func performFloatOp(l, r float64, op *Op, ctx *EvaluationContext) error {
	switch op.code {
	case opcAdd:
		ctx.Slots[*op.Out] = core.Num(l + r)
	case opcSub:
		ctx.Slots[*op.Out] = core.Num(l - r)
	case opcMul:
		ctx.Slots[*op.Out] = core.Num(l * r)
	case opcDiv:
		if r == 0 {
			ctx.Slots[*op.Out] = core.U() // div by zero is undefined
			return nil
		}
		ctx.Slots[*op.Out] = core.Num(l / r)
	case opcMod:
		if r == 0 {
			ctx.Slots[*op.Out] = core.U() // mod (div) by zero is undefined
			return nil
		}
		ctx.Slots[*op.Out] = core.Num(float64(int64(l) % int64(r)))
	case opcLt:
		ctx.Slots[*op.Out] = core.Bool(l < r)
	case opcGt:
		ctx.Slots[*op.Out] = core.Bool(l > r)
	case opcLte:
		ctx.Slots[*op.Out] = core.Bool(l <= r)
	case opcGte:
		ctx.Slots[*op.Out] = core.Bool(l >= r)
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

	switch op.code {
	case opcAnd:
		ctx.Slots[*op.Out] = kleeneAnd(lval, rval)
		return nil
	case opcOr:
		ctx.Slots[*op.Out] = kleeneOr(lval, rval)
		return nil
	}

	// if either side is undefined, result is undefined
	if lval.IsUndefined() || rval.IsUndefined() {
		ctx.Slots[*op.Out] = core.U()
		return nil
	}

	// handle operations that work on any type
	switch op.code {
	case opcEq:
		ctx.Slots[*op.Out] = core.Bool(core.SafeEquals(lval, rval))
		return nil
	case opcNeq:
		ctx.Slots[*op.Out] = core.Bool(!core.SafeEquals(lval, rval))
		return nil
	case opcIn:
		arr, ok := rval.AsArray()
		if !ok {
			return fmt.Errorf("right operand of '%v' must be an array, got %v", OpIn, rval.Type())
		}
		ctx.Slots[*op.Out] = core.Bool(contains(arr, lval))
		return nil
	}

	// handle string operations
	if l, ok := lval.AsString(); ok {
		switch op.code {
		case opcAdd:
			r, ok := rval.AsString()
			if !ok {
				ctx.Slots[*op.Out] = core.U()
				return nil
			}
			ctx.Slots[*op.Out] = core.V(l + r)
			return nil
		default:
			ctx.Slots[*op.Out] = core.U()
			return nil
		}
	}

	l, lok := lval.AsFloat()
	r, rok := rval.AsFloat()
	if !lok || !rok {
		ctx.Slots[*op.Out] = core.U()
		return nil
	}
	return performFloatOp(l, r, op, ctx)
}

// three-valued logical AND; false dominates (`false && undefined` is false)
func kleeneAnd(lval, rval core.Value) core.Value {
	lb, lok := lval.AsBool()
	rb, rok := rval.AsBool()
	if (lok && !lb) || (rok && !rb) {
		return core.Bool(false)
	}
	if lval.IsUndefined() || rval.IsUndefined() {
		return core.U()
	}
	if lok && rok {
		return core.Bool(lb && rb)
	}
	return core.U()
}

// three-valued logical OR; true dominates (`true || undefined` is true)
func kleeneOr(lval, rval core.Value) core.Value {
	lb, lok := lval.AsBool()
	rb, rok := rval.AsBool()
	if (lok && lb) || (rok && rb) {
		return core.Bool(true)
	}
	if lval.IsUndefined() || rval.IsUndefined() {
		return core.U()
	}
	if lok && rok {
		return core.Bool(lb || rb)
	}
	return core.U()
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
