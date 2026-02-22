package passes

import (
	"testing"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/nir"
)

func TestConstantPropagation(t *testing.T) {
	p := &nir.Program{
		Rules: map[string]nir.Rule{
			"Test": {
				Ops: []nir.Op{
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: core.V(42)}},
						},
						Out: core.IntPtr(1),
					},
					{
						Kind: nir.OpAdd,
						Args: []nir.RawExpr{
							{Expr: &nir.SymExpr{Index: 1}},
							{Expr: &nir.ImmExpr{Value: core.V(1)}},
						},
						Out: core.IntPtr(2),
					},
				},
			},
		},
	}

	cp := NewConstantPropagation()
	cp.Run(p)

	rule := p.Rules["Test"]
	op := rule.Ops[1]

	imm, ok := op.Args[0].Expr.(*nir.ImmExpr)
	if !ok {
		t.Errorf("Expected ImmExpr, but got %T", op.Args[0].Expr)
	}

	if !core.SafeEquals(imm.Value, core.V(42)) {
		t.Errorf("Expected ImmExpr to have value of 1, but got %v", imm.Value)
	}
}
