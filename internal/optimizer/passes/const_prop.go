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

func (cp *ConstantPropagation) Run(p *nir.Ruleset) (bool, error) {
	changed := false
	for ruleName, rule := range p.Rules {
		cp.constants = make(map[int]*nir.RawExpr)
		for i := range rule.Ops {
			op := &rule.Ops[i]

			if op.Kind == nir.OpCopy && cp.isConstant(op.Args[0]) {
				cp.constants[*op.Out] = &op.Args[0]
				continue
			}

			for i := range op.Args {
				if cp.propagateExpr(&op.Args[i]) {
					changed = true
				}
			}

			// re-check after propagation, args may now be fully constant
			if op.Kind == nir.OpCopy && op.Out != nil && cp.isConstant(op.Args[0]) {
				cp.constants[*op.Out] = &op.Args[0]
			}
		}
		p.Rules[ruleName] = rule
	}
	return changed, nil
}

func (cp *ConstantPropagation) propagateExpr(expr *nir.RawExpr) bool {
	switch e := expr.Expr.(type) {
	case *nir.SymExpr:
		if c, exists := cp.constants[e.Index]; exists {
			*expr = *c
			return true
		}
		return false
	case *nir.CallExpr:
		changed := false
		for i := range e.Args {
			if cp.propagateExpr(&e.Args[i]) {
				changed = true
			}
		}
		return changed
	case *nir.ArrExpr:
		changed := false
		for i := range e.Value {
			if cp.propagateExpr(&e.Value[i]) {
				changed = true
			}
		}
		return changed
	case *nir.ObjExpr:
		changed := false
		for k := range e.Value {
			v := e.Value[k]
			if cp.propagateExpr(&v) {
				changed = true
			}
			e.Value[k] = v
		}
		return changed
	case *nir.IdxExpr:
		c1 := cp.propagateExpr(&e.From)
		c2 := cp.propagateExpr(&e.Index)
		return c1 || c2
	case *nir.SelExpr:
		changed := cp.propagateExpr(&e.From)
		for i := range e.Exprs {
			if cp.propagateExpr(&e.Exprs[i]) {
				changed = true
			}
		}
		return changed
	case *nir.SignalExpr:
		changed := false
		for i := range e.Args {
			if cp.propagateExpr(&e.Args[i]) {
				changed = true
			}
		}
		if e.When != nil {
			if cp.propagateExpr(e.When) {
				changed = true
			}
		}
		return changed
	case *nir.LambdaExpr:
		changed := false
		for i := range e.Ops {
			for j := range e.Ops[i].Args {
				if cp.propagateExpr(&e.Ops[i].Args[j]) {
					changed = true
				}
			}
		}
		return changed
	default:
		return false
	}
}

func (cp *ConstantPropagation) isConstant(expr nir.RawExpr) bool {
	if expr.Expr == nil {
		return false
	}
	switch e := expr.Expr.(type) {
	case *nir.ImmExpr:
		return true
	case *nir.ArrExpr:
		for _, v := range e.Value {
			if !cp.isConstant(v) {
				return false
			}
		}
		return true
	case *nir.ObjExpr:
		for _, v := range e.Value {
			if !cp.isConstant(v) {
				return false
			}
		}
		return true
	default:
		return false
	}
}
