package passes

import (
	"testing"

	"nodora.org/nodora/pkg/nir"
)

func TestSymbolRemap_BasicRemapping(t *testing.T) {
	tests := []struct {
		name             string
		rule             nir.Rule
		expectedSlots    map[int]bool // which slots should exist after remap
		expectedSymslots int
	}{
		{
			name: "remaps sequential slots",
			rule: nir.Rule{
				Outputs: map[string]nir.Output{
					"result": {Sym: 3},
				},
				Ops: []nir.Op{
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: nir.V(1)}},
						},
						Out: nir.IntPtr(1),
					},
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.SymExpr{Index: 1}},
						},
						Out: nir.IntPtr(2),
					},
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.SymExpr{Index: 2}},
						},
						Out: nir.IntPtr(3),
					},
				},
				Symslots: 4,
			},
			expectedSlots:    map[int]bool{1: true, 2: true, 3: true},
			expectedSymslots: 4,
		},
		{
			name: "remaps sparse slots",
			rule: nir.Rule{
				Outputs: map[string]nir.Output{
					"result": {Sym: 10},
				},
				Ops: []nir.Op{
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: nir.V(1)}},
						},
						Out: nir.IntPtr(5),
					},
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.SymExpr{Index: 5}},
						},
						Out: nir.IntPtr(10),
					},
				},
				Symslots: 11,
			},
			expectedSlots:    map[int]bool{1: true, 2: true},
			expectedSymslots: 3,
		},
		{
			name: "handles operations without outputs",
			rule: nir.Rule{
				Outputs: map[string]nir.Output{},
				Ops: []nir.Op{
					{
						Kind: nir.OpEmit,
						Args: []nir.RawExpr{
							{Expr: &nir.SignalExpr{Name: "test"}},
						},
						Out: nil,
					},
				},
			},
			expectedSlots:    map[int]bool{},
			expectedSymslots: 1,
		},
		{
			name: "handles empty ops",
			rule: nir.Rule{
				Outputs: map[string]nir.Output{
					"output": {Sym: 0},
				},
				Ops:      []nir.Op{},
				Symslots: 1,
			},
			expectedSlots:    map[int]bool{},
			expectedSymslots: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &nir.Program{
				Rules: map[string]nir.Rule{
					"Test": tt.rule,
				},
			}

			sr := NewSymbolRemap()
			err := sr.Run(p)
			if err != nil {
				t.Fatalf("Run() returned an error: %v", err)
			}

			rule := p.Rules["Test"]

			// Check symslots
			if rule.Symslots != tt.expectedSymslots {
				t.Errorf("Expected Symslots=%d, got %d", tt.expectedSymslots, rule.Symslots)
			}

			// Collect all slot indices used in ops
			foundSlots := make(map[int]bool)
			for _, op := range rule.Ops {
				if op.Out != nil {
					foundSlots[*op.Out] = true
				}
			}

			// Check that slots match expected
			if len(foundSlots) != len(tt.expectedSlots) {
				t.Errorf("Expected %d slots, found %d", len(tt.expectedSlots), len(foundSlots))
			}
			for slot := range tt.expectedSlots {
				if !foundSlots[slot] {
					t.Errorf("Expected slot %d to exist", slot)
				}
			}
		})
	}
}

func TestSymbolRemap_SymExprRemapping(t *testing.T) {
	tests := []struct {
		name               string
		rule               nir.Rule
		expectedSymIndices map[int]int // old index -> new index
		expectedSymslots   int
	}{
		{
			name: "remaps SymExpr in args",
			rule: nir.Rule{
				Symslots: 6,
				Outputs: map[string]nir.Output{
					"result": {Sym: 2},
				},
				Ops: []nir.Op{
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: nir.V(42)}},
						},
						Out: nir.IntPtr(5),
					},
					{
						Kind: nir.OpAdd,
						Args: []nir.RawExpr{
							{Expr: &nir.SymExpr{Index: 5}},
							{Expr: &nir.ImmExpr{Value: nir.V(10)}},
						},
						Out: nir.IntPtr(2),
					},
				},
			},
			expectedSymIndices: map[int]int{5: 1, 2: 2},
			expectedSymslots:   3,
		},
		{
			name: "remaps multiple SymExprs",
			rule: nir.Rule{
				Symslots: 11,
				Outputs: map[string]nir.Output{
					"result": {Sym: 10},
				},
				Ops: []nir.Op{
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: nir.V(1)}},
						},
						Out: nir.IntPtr(1),
					},
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: nir.V(2)}},
						},
						Out: nir.IntPtr(3),
					},
					{
						Kind: nir.OpAdd,
						Args: []nir.RawExpr{
							{Expr: &nir.SymExpr{Index: 1}},
							{Expr: &nir.SymExpr{Index: 3}},
						},
						Out: nir.IntPtr(10),
					},
				},
			},
			expectedSymIndices: map[int]int{1: 1, 3: 2, 10: 3},
			expectedSymslots:   4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &nir.Program{
				Rules: map[string]nir.Rule{
					"Test": tt.rule,
				},
			}

			sr := NewSymbolRemap()
			err := sr.Run(p)
			if err != nil {
				t.Fatalf("Run() returned an error: %v", err)
			}

			rule := p.Rules["Test"]

			// Check that SymExpr indices are remapped correctly
			for _, op := range rule.Ops {
				for _, arg := range op.Args {
					sr.checkSymExpr(t, &arg, tt.expectedSymslots)
				}
			}

			// Check symslots
			if rule.Symslots != tt.expectedSymslots {
				t.Errorf("Expected Symslots=%d, got %d", tt.expectedSymslots, rule.Symslots)
			}

			// Check outputs are remapped
			for name, output := range rule.Outputs {
				found := false
				for oldIdx, newIdx := range tt.expectedSymIndices {
					if output.Sym == newIdx {
						found = true
						// The output should reference one of the new indices
						_ = oldIdx
						break
					}
				}
				if !found && len(tt.expectedSymIndices) > 0 {
					t.Errorf("Output %s has unmapped Sym index %d", name, output.Sym)
				}
			}
		})
	}
}

func (sr *SymbolRemap) checkSymExpr(t *testing.T, expr *nir.RawExpr, expectedSymslots int) {
	switch e := expr.Expr.(type) {
	case *nir.SymExpr:
		if e.Index < 0 || e.Index >= expectedSymslots {
			t.Errorf("SymExpr has invalid index %d, expected < %v", e.Index, expectedSymslots)
		}
	case *nir.IdxExpr:
		sr.checkSymExpr(t, &e.Index, expectedSymslots)
		sr.checkSymExpr(t, &e.From, expectedSymslots)
	case *nir.SelExpr:
		sr.checkSymExpr(t, &e.From, expectedSymslots)
		for _, ex := range e.Exprs {
			sr.checkSymExpr(t, &ex, expectedSymslots)
		}
	case *nir.ArrExpr:
		for _, v := range e.Value {
			sr.checkSymExpr(t, &v, expectedSymslots)
		}
	case *nir.SignalExpr:
		for _, arg := range e.Args {
			sr.checkSymExpr(t, &arg, expectedSymslots)
		}
		if e.When != nil {
			sr.checkSymExpr(t, e.When, expectedSymslots)
		}
	}
}

func TestSymbolRemap_ComplexExpressions(t *testing.T) {
	tests := []struct {
		name  string
		rule  nir.Rule
		check func(t *testing.T, rule nir.Rule)
	}{
		{
			name: "remaps SymExpr in IdxExpr",
			rule: nir.Rule{
				Symslots: 6,
				Outputs: map[string]nir.Output{
					"result": {Sym: 2},
				},
				Ops: []nir.Op{
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ArrExpr{Value: []nir.RawExpr{
								{Expr: &nir.ImmExpr{Value: nir.V(1)}},
							}}},
						},
						Out: nir.IntPtr(5),
					},
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{
								Expr: &nir.IdxExpr{
									Index: nir.RawExpr{
										Expr: &nir.ImmExpr{Value: nir.V(0)},
									},
									From: nir.RawExpr{
										Expr: &nir.SymExpr{Index: 5},
									},
								},
							},
						},
						Out: nir.IntPtr(2),
					},
				},
			},
			check: func(t *testing.T, rule nir.Rule) {
				// Check that SymExpr inside IdxExpr was remapped
				op := rule.Ops[1]
				idxExpr := op.Args[0].Expr.(*nir.IdxExpr)
				symExpr := idxExpr.From.Expr.(*nir.SymExpr)
				if symExpr.Index != 1 {
					t.Errorf("Expected SymExpr index 1, got %d", symExpr.Index)
				}
			},
		},
		{
			name: "remaps SymExpr in SelExpr",
			rule: nir.Rule{
				Outputs: map[string]nir.Output{
					"result": {Sym: 2},
				},
				Ops: []nir.Op{
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: nir.V(nir.ValueMap{"field": nir.V(1)})}},
						},
						Out: nir.IntPtr(5),
					},
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{
								Expr: &nir.SelExpr{
									Path: "field",
									From: nir.RawExpr{Expr: &nir.SymExpr{Index: 5}},
								},
							},
						},
						Out: nir.IntPtr(2),
					},
				},
			},
			check: func(t *testing.T, rule nir.Rule) {
				// Check that SymExpr inside SelExpr was remapped
				op := rule.Ops[1]
				selExpr := op.Args[0].Expr.(*nir.SelExpr)
				symExpr := selExpr.From.Expr.(*nir.SymExpr)
				if symExpr.Index != 1 {
					t.Errorf("Expected SymExpr index 1, got %d", symExpr.Index)
				}
			},
		},
		{
			name: "remaps SymExpr in ArrExpr",
			rule: nir.Rule{
				Outputs: map[string]nir.Output{
					"result": {Sym: 2},
				},
				Ops: []nir.Op{
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: nir.V(100)}},
						},
						Out: nir.IntPtr(5),
					},
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{
								Expr: &nir.ArrExpr{
									Value: []nir.RawExpr{
										{Expr: &nir.SymExpr{Index: 5}},
									},
								},
							},
						},
						Out: nir.IntPtr(2),
					},
				},
			},
			check: func(t *testing.T, rule nir.Rule) {
				// Check that SymExpr inside ArrExpr was remapped
				op := rule.Ops[1]
				arrExpr := op.Args[0].Expr.(*nir.ArrExpr)
				symExpr := arrExpr.Value[0].Expr.(*nir.SymExpr)
				if symExpr.Index != 1 {
					t.Errorf("Expected SymExpr index 1, got %d", symExpr.Index)
				}
			},
		},
		{
			name: "remaps SymExpr in SignalExpr args and When",
			rule: nir.Rule{
				Outputs: map[string]nir.Output{
					"result": {Sym: 3},
				},
				Ops: []nir.Op{
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: nir.V(100)}},
						},
						Out: nir.IntPtr(5),
					},
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: nir.V(true)}},
						},
						Out: nir.IntPtr(10),
					},
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{
								Expr: &nir.SignalExpr{
									Name: "test",
									Args: []nir.RawExpr{
										{Expr: &nir.SymExpr{Index: 5}},
									},
									When: &nir.RawExpr{Expr: &nir.SymExpr{Index: 10}},
								},
							},
						},
						Out: nir.IntPtr(3),
					},
				},
			},
			check: func(t *testing.T, rule nir.Rule) {
				// Check that SymExprs inside SignalExpr were remapped
				op := rule.Ops[2]
				sigExpr := op.Args[0].Expr.(*nir.SignalExpr)
				argSymExpr := sigExpr.Args[0].Expr.(*nir.SymExpr)
				whenSymExpr := sigExpr.When.Expr.(*nir.SymExpr)
				if argSymExpr.Index != 1 {
					t.Errorf("Expected SignalExpr arg SymExpr index 1, got %d", argSymExpr.Index)
				}
				if whenSymExpr.Index != 2 {
					t.Errorf("Expected SignalExpr When SymExpr index 2, got %d", whenSymExpr.Index)
				}
			},
		},
		{
			name: "handles nested expressions",
			rule: nir.Rule{
				Outputs: map[string]nir.Output{
					"result": {Sym: 2},
				},
				Ops: []nir.Op{
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: nir.V(100)}},
						},
						Out: nir.IntPtr(5),
					},
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{
								Expr: &nir.IdxExpr{
									Index: nir.RawExpr{Expr: &nir.ImmExpr{Value: nir.V(0)}},
									From: nir.RawExpr{Expr: &nir.ArrExpr{
										Value: []nir.RawExpr{
											{Expr: &nir.SymExpr{Index: 5}},
										},
									}},
								},
							},
						},
						Out: nir.IntPtr(2),
					},
				},
			},
			check: func(t *testing.T, rule nir.Rule) {
				// Check that deeply nested SymExpr was remapped
				op := rule.Ops[1]
				idxExpr := op.Args[0].Expr.(*nir.IdxExpr)
				arrExpr := idxExpr.From.Expr.(*nir.ArrExpr)
				symExpr := arrExpr.Value[0].Expr.(*nir.SymExpr)
				if symExpr.Index != 1 {
					t.Errorf("Expected nested SymExpr index 1, got %d", symExpr.Index)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &nir.Program{
				Rules: map[string]nir.Rule{
					"Test": tt.rule,
				},
			}

			sr := NewSymbolRemap()
			err := sr.Run(p)
			if err != nil {
				t.Fatalf("Run() returned an error: %v", err)
			}

			rule := p.Rules["Test"]
			tt.check(t, rule)
		})
	}
}

func TestSymbolRemap_MultipleRules(t *testing.T) {
	p := &nir.Program{
		Rules: map[string]nir.Rule{
			"Rule1": {
				Outputs: map[string]nir.Output{
					"out": {Sym: 5},
				},
				Ops: []nir.Op{
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: nir.V(1)}},
						},
						Out: nir.IntPtr(5),
					},
				},
				Symslots: 6,
			},
			"Rule2": {
				Outputs: map[string]nir.Output{
					"out": {Sym: 10},
				},
				Ops: []nir.Op{
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: nir.V(2)}},
						},
						Out: nir.IntPtr(10),
					},
				},
				Symslots: 11,
			},
		},
	}

	sr := NewSymbolRemap()
	err := sr.Run(p)
	if err != nil {
		t.Fatalf("Run() returned an error: %v", err)
	}

	// Check Rule1
	rule1 := p.Rules["Rule1"]
	if rule1.Symslots != 2 {
		t.Errorf("Rule1: Expected Symslots=2, got %d", rule1.Symslots)
	}
	if rule1.Ops[0].Out == nil || *rule1.Ops[0].Out != 1 {
		t.Errorf("Rule1: Expected output slot 1, got %v", rule1.Ops[0].Out)
	}
	if rule1.Outputs["out"].Sym != 1 {
		t.Errorf("Rule1: Expected output Sym=1, got %d", rule1.Outputs["out"].Sym)
	}

	// Check Rule2
	rule2 := p.Rules["Rule2"]
	if rule2.Symslots != 2 {
		t.Errorf("Rule2: Expected Symslots=2, got %d", rule2.Symslots)
	}
	if rule2.Ops[0].Out == nil || *rule2.Ops[0].Out != 1 {
		t.Errorf("Rule2: Expected output slot 1, got %v", rule2.Ops[0].Out)
	}
	if rule2.Outputs["out"].Sym != 1 {
		t.Errorf("Rule2: Expected output Sym=1, got %d", rule2.Outputs["out"].Sym)
	}
}

func TestSymbolRemap_NilHandling(t *testing.T) {
	// Test that nil expressions don't cause panics
	rule := nir.Rule{
		Outputs: map[string]nir.Output{},
		Ops: []nir.Op{
			{
				Kind: nir.OpEmit,
				Args: []nir.RawExpr{
					{Expr: &nir.SignalExpr{
						Name: "test",
						Args: []nir.RawExpr{},
						When: nil, // nil When clause
					}},
				},
				Out: nil,
			},
		},
	}

	p := &nir.Program{
		Rules: map[string]nir.Rule{
			"Test": rule,
		},
	}

	sr := NewSymbolRemap()
	err := sr.Run(p)
	if err != nil {
		t.Fatalf("Run() returned an error: %v", err)
	}

	// Just verify it doesn't panic
	r := p.Rules["Test"]
	if len(r.Ops) != 1 {
		t.Errorf("Expected 1 op, got %d", len(r.Ops))
	}
}
