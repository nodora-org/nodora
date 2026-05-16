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
	_, _ = cp.Run(p)

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

// ensures that constants from one rule don't leak into another
func TestConstantPropagationRuleIsolation(t *testing.T) {
	p := &nir.Program{
		Rules: map[string]nir.Rule{
			"RuleA": {
				Ops: []nir.Op{
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{{Expr: &nir.ImmExpr{Value: core.V(false)}}},
						Out:  core.IntPtr(1),
					},
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{{Expr: &nir.ImmExpr{Value: core.V(float64(42))}}},
						Out:  core.IntPtr(2),
					},
				},
			},
			"RuleB": {
				Ops: []nir.Op{
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{{Expr: &nir.SelExpr{
							Path: "a",
							From: nir.RawExpr{Expr: &nir.SymExpr{Index: 0}},
						}}},
						Out: core.IntPtr(1),
					},
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{{Expr: &nir.SelExpr{
							Path: "b",
							From: nir.RawExpr{Expr: &nir.SymExpr{Index: 0}},
						}}},
						Out: core.IntPtr(2),
					},
					{
						Kind: nir.OpAnd,
						Args: []nir.RawExpr{
							{Expr: &nir.SymExpr{Index: 1}},
							{Expr: &nir.SymExpr{Index: 2}},
						},
						Out: core.IntPtr(3),
					},
				},
			},
		},
	}

	rp := NewRepeatedPass(10, NewConstantFolding(), NewConstantPropagation())
	if _, err := rp.Run(p); err != nil {
		t.Errorf("optimizer returned error due to cross-rule constant leak: %v", err)
	}
}