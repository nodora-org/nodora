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

	// check if all arguments are constants
	for i := range op.Args {
		if folded, ok := cf.foldExpr(op.Args[i]); ok {
			op.Args[i] = *folded
		}
		if !cf.isConst(op.Args[i]) {
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

func (cf *ConstantFolding) foldExpr(expr nir.RawExpr) (*nir.RawExpr, bool) {
	if !cf.isConst(expr) {
		return nil, false
	}
	val, err := expr.Evaluate(&nir.EvaluationContext{})
	if err != nil {
		return nil, false
	}
	immExpr := nir.RawExpr{Expr: &nir.ImmExpr{Value: val}}
	return &immExpr, true
}

func (cf *ConstantFolding) isConst(expr nir.RawExpr) bool {
	if expr.Expr == nil {
		return false
	}
	switch e := expr.Expr.(type) {
	case *nir.ImmExpr:
		return true
	case *nir.ArrExpr:
		for _, arg := range e.Value {
			if !cf.isConst(arg) {
				return false
			}
		}
		return true
	case *nir.ObjExpr:
		for _, value := range e.Value {
			if !cf.isConst(value) {
				return false
			}
		}
		return true
	case *nir.IdxExpr:
		return cf.isConst(e.From) && cf.isConst(e.Index)
	case *nir.CallExpr:
		for _, arg := range e.Args {
			if !cf.isConst(arg) {
				return false
			}
		}
		return true
	default:
		return false
	}
}
