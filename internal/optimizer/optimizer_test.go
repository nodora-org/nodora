package optimizer

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"nodora.org/nodora/internal/optimizer/passes"
	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/nir"
)

func TestOptimizer(t *testing.T) {
	ruleset := nir.Ruleset{
		Signals: map[string]nir.Signal{
			"Signal1": {Params: []nir.Param{{Name: "x"}}},
			"Signal2": {Params: []nir.Param{{Name: "y"}}},
			"Signal3": {Params: []nir.Param{{Name: "z"}}},
		},
		Rules: map[string]nir.Rule{
			"Rule1": {
				Symslots: 5,
				Outputs: map[string]nir.Output{
					"a": {Sym: 3},
					"b": {Sym: 4},
				},
				Ops: []nir.Op{
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: core.V(int64(1))}},
						},
						Out: core.IntPtr(1), // slot1 = 1 (unused)
					},
					{
						Kind: nir.OpAdd,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: core.V(int64(2))}},
							{Expr: &nir.ImmExpr{Value: core.V(int64(2))}},
						},
						Out: core.IntPtr(2), // slot2 = 2 + 2
					},
					{
						Kind: nir.OpMul,
						Args: []nir.RawExpr{
							{Expr: &nir.SymExpr{Index: 2}},
							{Expr: &nir.SymExpr{Index: 2}},
						},
						Out: core.IntPtr(3), // a = slot2 * slot2
					},
					{
						Kind: nir.OpLte,
						Args: []nir.RawExpr{
							{Expr: &nir.SymExpr{Index: 3}},
							{Expr: &nir.ImmExpr{Value: core.V(int64(16))}},
						},
						Out: core.IntPtr(4), // b = a <= 16
					},
					{
						Kind: nir.OpEmit,
						Args: []nir.RawExpr{
							{Expr: &nir.SignalExpr{
								Name: "Signal1",
								Args: []nir.RawExpr{
									{Expr: &nir.SymExpr{Index: 3}},
								},
								When: &nir.RawExpr{
									Expr: &nir.SymExpr{Index: 4},
								},
							}},
						},
						Out: nil, // emit Signal1(a) when b
					},
				},
			},
			"Rule2": {
				Symslots: 3,
				Outputs:  map[string]nir.Output{},
				Ops: []nir.Op{
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{
								Expr: &nir.SelExpr{
									Path: "field",
									From: nir.RawExpr{Expr: &nir.SymExpr{Index: 0}},
								},
							},
						},
						Out: core.IntPtr(1),
					},
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.SymExpr{Index: 1}},
						},
						Out: core.IntPtr(2),
					},
					{
						Kind: nir.OpEmit,
						Args: []nir.RawExpr{
							{Expr: &nir.SignalExpr{
								Name: "Signal2",
								Args: []nir.RawExpr{
									{Expr: &nir.ImmExpr{Value: core.V(int64(1))}},
								},
							}},
						},
						Out: nil,
					},
				},
			},
		},
	}

	optimizer := NewOptimizer()
	optimizer.AddPass(passes.NewConstantFolding())
	optimizer.AddPass(passes.NewDeadCodeElimination())
	optimizer.AddPass(passes.NewSymbolRemap())

	if err := optimizer.Run(&ruleset); err != nil {
		t.Fatalf("Optimizer returned an error: %v", err)
	}

	expectedRuleset := nir.Ruleset{
		Signals: map[string]nir.Signal{
			"Signal1": {Params: []nir.Param{{Name: "x"}}},
			"Signal2": {Params: []nir.Param{{Name: "y"}}},
		},
		Rules: map[string]nir.Rule{
			"Rule1": {
				Symslots: 4,
				Outputs: map[string]nir.Output{
					"a": {Sym: 2},
					"b": {Sym: 3},
				},
				Ops: []nir.Op{
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: core.V(int64(4))}},
						},
						Out: core.IntPtr(1), // slot1 = 4
					},
					{
						Kind: nir.OpMul,
						Args: []nir.RawExpr{
							{Expr: &nir.SymExpr{Index: 1}},
							{Expr: &nir.SymExpr{Index: 1}},
						},
						Out: core.IntPtr(2), // a = slot1 * slot1
					},
					{
						Kind: nir.OpLte,
						Args: []nir.RawExpr{
							{Expr: &nir.SymExpr{Index: 2}},
							{Expr: &nir.ImmExpr{Value: core.V(int64(16))}},
						},
						Out: core.IntPtr(3), // b = a <= 16
					},
					{
						Kind: nir.OpEmit,
						Args: []nir.RawExpr{
							{Expr: &nir.SignalExpr{
								Name: "Signal1",
								Args: []nir.RawExpr{
									{Expr: &nir.SymExpr{Index: 2}},
								},
								When: &nir.RawExpr{
									Expr: &nir.SymExpr{Index: 3},
								},
							}},
						},
						Out: nil, // emit Signal1(a) when b
					},
				},
			},
			"Rule2": {
				Symslots: 1,
				Outputs:  map[string]nir.Output{},
				Ops: []nir.Op{
					{
						Kind: nir.OpEmit,
						Args: []nir.RawExpr{
							{Expr: &nir.SignalExpr{
								Name: "Signal2",
								Args: []nir.RawExpr{
									{Expr: &nir.ImmExpr{Value: core.V(int64(1))}},
								},
							}},
						},
						Out: nil,
					},
				},
			},
		},
	}

	// ignore unexported state
	opts := cmpopts.IgnoreUnexported(nir.Op{}, nir.CallExpr{}, nir.SelExpr{})
	if diff := cmp.Diff(expectedRuleset, ruleset, opts); diff != "" {
		t.Errorf("ruleset mismatch (-expected +got):\n%s", diff)
	}
}
