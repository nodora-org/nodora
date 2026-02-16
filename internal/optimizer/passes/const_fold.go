package passes

import (
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
	if op.Out == nil || op.Kind == nir.OpCopy {
		return nil
	}

	// check if all arguments are constants
	for _, arg := range op.Args {
		if !cf.isConstant(arg) {
			return nil
		}
	}

	evalCtx := nir.EvaluationContext{
		Slots: make([]nir.Value, *op.Out+1),
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

func (cf *ConstantFolding) isConstant(expr nir.RawExpr) bool {
	if expr.Expr == nil {
		return false
	}
	_, ok := expr.Expr.(*nir.ImmExpr)
	return ok
}
