package passes

import (
	"nodora.org/nodora/pkg/nir"
)

type ConstantPropagation struct {
	constants map[int]*nir.RawExpr
}

func NewConstantPropagation() *ConstantPropagation {
	return &ConstantPropagation{
		constants: make(map[int]*nir.RawExpr),
	}
}

func (cp *ConstantPropagation) Run(p *nir.Program) error {
	for ruleName, rule := range p.Rules {
		for i := range rule.Ops {
			op := &rule.Ops[i]

			if op.Kind == nir.OpCopy && cp.isConstant(op.Args[0]) {
				cp.constants[*op.Out] = &op.Args[0]
				continue
			}

			for i := range op.Args {
				cp.propagateExpr(&op.Args[i])
			}
		}
		p.Rules[ruleName] = rule
	}
	return nil
}

func (cp *ConstantPropagation) propagateExpr(expr *nir.RawExpr) {
	switch e := expr.Expr.(type) {
	case *nir.SymExpr:
		if c, exists := cp.constants[e.Index]; exists {
			*expr = *c
		}
	case *nir.CallExpr:
		for i := range e.Args {
			cp.propagateExpr(&e.Args[i])
		}
	case *nir.ArrExpr:
		for i := range e.Value {
			cp.propagateExpr(&e.Value[i])
		}
	case *nir.ObjExpr:
		for k := range e.Value {
			v := e.Value[k]
			cp.propagateExpr(&v)
			e.Value[k] = v
		}
	case *nir.IdxExpr:
		cp.propagateExpr(&e.From)
		cp.propagateExpr(&e.Index)
	case *nir.SelExpr:
		cp.propagateExpr(&e.From)
		for i := range e.Exprs {
			cp.propagateExpr(&e.Exprs[i])
		}
	case *nir.SignalExpr:
		for i := range e.Args {
			cp.propagateExpr(&e.Args[i])
		}
		if e.When != nil {
			cp.propagateExpr(e.When)
		}
	case *nir.LambdaExpr:
		for i := range e.Ops {
			for j := range e.Ops[i].Args {
				cp.propagateExpr(&e.Ops[i].Args[j])
			}
		}
	}
}

func (cp *ConstantPropagation) isConstant(expr nir.RawExpr) bool {
	if expr.Expr == nil {
		return false
	}
	switch expr.Expr.(type) {
	case *nir.ImmExpr:
		return true
	default:
		return false
	}
}
