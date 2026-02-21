package passes

import (
	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/nir"
)

type ConstantFolding struct{}

func NewConstantFolding() *ConstantFolding {
	return &ConstantFolding{}
}

func (cf *ConstantFolding) Run(p *nir.Program) error {
	for ruleName, rule := range p.Rules {
		for i := range rule.Ops {
			if err := cf.foldOp(&rule.Ops[i]); err != nil {
				return err
			}
		}
		p.Rules[ruleName] = rule
	}
	return nil
}

func (cf *ConstantFolding) foldOp(op *nir.Op) error {
	if op.Out == nil {
		return nil
	}

	// try to fold any len() calls first
	for i, arg := range op.Args {
		if folded, ok := cf.foldLenCall(arg); ok {
			op.Args[i] = *folded
		}
	}

	// check if all arguments are constants
	for _, arg := range op.Args {
		if !cf.isConstant(arg) {
			return nil
		}
	}

	evalCtx := nir.EvaluationContext{
		Slots: make([]core.Value, *op.Out+1),
	}

	// execute the op
	if err := op.Execute(&evalCtx); err != nil {
		return err
	}

	result := evalCtx.Slots[*op.Out]

	// replace with a copy op containing the constant result
	op.Kind = nir.OpCopy
	op.Args = []nir.RawExpr{{Expr: &nir.ImmExpr{Value: result}}}

	return nil
}

// folds len() calls on array literals
func (cf *ConstantFolding) foldLenCall(expr nir.RawExpr) (*nir.RawExpr, bool) {
	callExpr, ok := expr.Expr.(*nir.CallExpr)
	if !ok {
		return nil, false
	}

	if !cf.isLenCall(callExpr) {
		return nil, false
	}

	arrExpr, ok := callExpr.Args[0].Expr.(*nir.ArrExpr)
	if !ok {
		return nil, false
	}

	resExpr := nir.RawExpr{Expr: &nir.ImmExpr{Value: core.V(len(arrExpr.Value))}}
	return &resExpr, true
}

func (cf *ConstantFolding) isConstant(expr nir.RawExpr) bool {
	if expr.Expr == nil {
		return false
	}
	switch e := expr.Expr.(type) {
	case *nir.ImmExpr:
		return true
	case *nir.ArrExpr:
		for _, arg := range e.Value {
			if !cf.isConstant(arg) {
				return false
			}
		}
		return true
	case *nir.ObjExpr:
		for _, value := range e.Value {
			if !cf.isConstant(value) {
				return false
			}
		}
		return true
	case *nir.CallExpr:
		for _, arg := range e.Args {
			if !cf.isConstant(arg) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (cf *ConstantFolding) isLenCall(e *nir.CallExpr) bool {
	return e.Func.Namespace == "" && e.Func.Name == "len"
}
