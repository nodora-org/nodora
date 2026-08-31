package passes

import (
	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/nir"
	"nodora.org/nodora/pkg/registry"
)

type ConstantFolding struct{}

func NewConstantFolding() *ConstantFolding {
	return &ConstantFolding{}
}

func (cf *ConstantFolding) Run(p *nir.Ruleset) (bool, error) {
	if err := p.Prepare(); err != nil {
		return false, err
	}

	changed := false
	for ruleName, rule := range p.Rules {
		for i := range rule.Ops {
			c, err := cf.foldOp(&rule.Ops[i])
			if err != nil {
				return changed, err
			}
			changed = changed || c
		}
		p.Rules[ruleName] = rule
	}
	return changed, nil
}

func (cf *ConstantFolding) foldOp(op *nir.Op) (bool, error) {
	if op.Out == nil {
		return false, nil
	}

	changed := false
	for i := range op.Args {
		if folded, ok := cf.foldExpr(op.Args[i]); ok {
			op.Args[i] = *folded
			changed = true
		}
	}
	for i := range op.Args {
		if !cf.isConst(op.Args[i]) {
			return changed, nil
		}
	}

	// execute the op in isolation (its args are all const here)
	evalCtx := nir.EvaluationContext{
		Slots: make([]core.Value, *op.Out+1),
	}
	if err := op.Execute(&evalCtx); err != nil {
		return false, err
	}
	result := evalCtx.Slots[*op.Out]

	// undefined result cannot be represented as an immediate:
	// leave the op to evaluate at runtime, where undefined is handled correctly
	if result.IsUndefined() {
		return changed, nil
	}

	// replace with a copy op containing the constant result
	op.Kind = nir.OpCopy
	op.Args = []nir.RawExpr{{Expr: &nir.ImmExpr{Value: result}}}

	return true, nil
}

func (cf *ConstantFolding) foldExpr(expr nir.RawExpr) (*nir.RawExpr, bool) {
	if !cf.isConst(expr) {
		return nil, false
	}
	val, err := expr.Evaluate(&nir.EvaluationContext{})
	if err != nil {
		return nil, false
	}
	if val.IsUndefined() {
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
	case *nir.SelExpr:
		return cf.isConst(e.From)
	case *nir.CallExpr:
		fn, ok := registry.Global().Get(e.Func.Namespace, e.Func.Name)
		if !ok || !fn.Pure {
			return false
		}
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
