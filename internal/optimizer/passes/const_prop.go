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
				arg := &op.Args[i]
				switch e := arg.Expr.(type) {
				case *nir.SymExpr:
					if _, exists := cp.constants[e.Index]; exists {
						op.Args[i] = *cp.constants[e.Index]
					}
				default:
					continue
				}
			}
		}
		p.Rules[ruleName] = rule
	}
	return nil
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
