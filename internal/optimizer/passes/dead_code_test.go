package passes

import (
	"testing"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/nir"
)

func TestDeadCodeElimination_UnusedSignals(t *testing.T) {
	tests := []struct {
		name            string
		signals         map[string]nir.Signal
		rules           map[string]nir.Rule
		expectedSignals []string
	}{
		{
			name: "removes unused signals",
			signals: map[string]nir.Signal{
				"usedSignal":   {Params: []nir.Param{{Name: "x"}}},
				"unusedSignal": {Params: []nir.Param{{Name: "y"}}},
			},
			rules: map[string]nir.Rule{
				"Test": {
					Ops: []nir.Op{
						{
							Kind: nir.OpEmit,
							Args: []nir.RawExpr{
								{Expr: &nir.SignalExpr{Name: "usedSignal"}},
							},
						},
					},
				},
			},
			expectedSignals: []string{"usedSignal"},
		},
		{
			name: "keeps all signals when all are used",
			signals: map[string]nir.Signal{
				"signal1": {Params: []nir.Param{{Name: "x"}}},
				"signal2": {Params: []nir.Param{{Name: "y"}}},
			},
			rules: map[string]nir.Rule{
				"Test1": {
					Ops: []nir.Op{
						{
							Kind: nir.OpEmit,
							Args: []nir.RawExpr{
								{Expr: &nir.SignalExpr{Name: "signal1"}},
							},
						},
					},
				},
				"Test2": {
					Ops: []nir.Op{
						{
							Kind: nir.OpEmit,
							Args: []nir.RawExpr{
								{Expr: &nir.SignalExpr{Name: "signal2"}},
							},
						},
					},
				},
			},
			expectedSignals: []string{"signal1", "signal2"},
		},
		{
			name: "removes all signals when none are used",
			signals: map[string]nir.Signal{
				"unused1": {Params: []nir.Param{{Name: "x"}}},
				"unused2": {Params: []nir.Param{{Name: "y"}}},
			},
			rules: map[string]nir.Rule{
				"Test": {
					Symslots: 1,
					Ops: []nir.Op{
						{
							Kind: nir.OpCopy,
							Args: []nir.RawExpr{
								{Expr: &nir.ImmExpr{Value: core.V(42)}},
							},
							Out: core.IntPtr(0),
						},
					},
				},
			},
			expectedSignals: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &nir.Program{
				Signals: tt.signals,
				Rules:   tt.rules,
			}

			dce := NewDeadCodeElimination()
			_, err := dce.Run(p)
			if err != nil {
				t.Fatalf("Run() returned error: %v", err)
			}

			if len(p.Signals) != len(tt.expectedSignals) {
				t.Errorf("Expected %d signals, got %d", len(tt.expectedSignals), len(p.Signals))
			}

			for _, expected := range tt.expectedSignals {
				if _, exists := p.Signals[expected]; !exists {
					t.Errorf("Expected signal %q to exist", expected)
				}
			}
		})
	}
}

func TestDeadCodeElimination_DeadOperations(t *testing.T) {
	tests := []struct {
		name            string
		rule            nir.Rule
		expectedOpCount int
		expectedOpKinds []nir.OpKind
	}{
		{
			name: "removes dead operations",
			rule: nir.Rule{
				Symslots: 2,
				Outputs: map[string]nir.Output{
					"result": {Sym: 1},
				},
				Ops: []nir.Op{
					{
						// dead op - result not used
						Kind: nir.OpAdd,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: core.V(1)}},
							{Expr: &nir.ImmExpr{Value: core.V(2)}},
						},
						Out: core.IntPtr(0),
					},
					{
						// live op - output 2 is used
						Kind: nir.OpAdd,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: core.V(3)}},
							{Expr: &nir.ImmExpr{Value: core.V(4)}},
						},
						Out: core.IntPtr(1),
					},
				},
			},
			expectedOpCount: 1,
			expectedOpKinds: []nir.OpKind{nir.OpAdd},
		},
		{
			name: "keeps operations with side effects",
			rule: nir.Rule{
				Outputs: map[string]nir.Output{},
				Ops: []nir.Op{
					{
						// side effect op (no Out) - should be kept
						Kind: nir.OpEmit,
						Args: []nir.RawExpr{
							{Expr: &nir.SignalExpr{Name: "testSignal"}},
						},
						Out: nil,
					},
				},
			},
			expectedOpCount: 1,
			expectedOpKinds: []nir.OpKind{nir.OpEmit},
		},
		{
			name: "keeps chained operations",
			rule: nir.Rule{
				Symslots: 3,
				Outputs: map[string]nir.Output{
					"result": {Sym: 2},
				},
				Ops: []nir.Op{
					{
						// slot 0 = 1 + 2
						Kind: nir.OpAdd,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: core.V(1)}},
							{Expr: &nir.ImmExpr{Value: core.V(2)}},
						},
						Out: core.IntPtr(0),
					},
					{
						// slot 1 = 3 + 4
						Kind: nir.OpAdd,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: core.V(3)}},
							{Expr: &nir.ImmExpr{Value: core.V(4)}},
						},
						Out: core.IntPtr(1),
					},
					{
						// slot 2 = slot 0 + slot 1
						Kind: nir.OpAdd,
						Args: []nir.RawExpr{
							{Expr: &nir.SymExpr{Index: 0}},
							{Expr: &nir.SymExpr{Index: 1}},
						},
						Out: core.IntPtr(2),
					},
				},
			},
			expectedOpCount: 3,
			expectedOpKinds: []nir.OpKind{nir.OpAdd, nir.OpAdd, nir.OpAdd},
		},
		{
			name: "removes operations with unused intermediate results",
			rule: nir.Rule{
				Symslots: 3,
				Outputs: map[string]nir.Output{
					"result": {Sym: 2},
				},
				Ops: []nir.Op{
					{
						// dead - slot 0 not used
						Kind: nir.OpAdd,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: core.V(1)}},
							{Expr: &nir.ImmExpr{Value: core.V(2)}},
						},
						Out: core.IntPtr(0),
					},
					{
						// slot 1 = 3 + 4
						Kind: nir.OpAdd,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: core.V(3)}},
							{Expr: &nir.ImmExpr{Value: core.V(4)}},
						},
						Out: core.IntPtr(1),
					},
					{
						// slot 2 = slot 1 + 5
						Kind: nir.OpAdd,
						Args: []nir.RawExpr{
							{Expr: &nir.SymExpr{Index: 1}},
							{Expr: &nir.ImmExpr{Value: core.V(5)}},
						},
						Out: core.IntPtr(2),
					},
				},
			},
			expectedOpCount: 2,
			expectedOpKinds: []nir.OpKind{nir.OpAdd, nir.OpAdd},
		},
		{
			name: "handles empty ops",
			rule: nir.Rule{
				Outputs: map[string]nir.Output{},
				Ops:     []nir.Op{},
			},
			expectedOpCount: 0,
			expectedOpKinds: []nir.OpKind{},
		},
		{
			name: "keeps multiple side effect operations",
			rule: nir.Rule{
				Outputs: map[string]nir.Output{},
				Ops: []nir.Op{
					{
						Kind: nir.OpEmit,
						Args: []nir.RawExpr{
							{Expr: &nir.SignalExpr{Name: "signal1"}},
						},
						Out: nil,
					},
					{
						Kind: nir.OpEmit,
						Args: []nir.RawExpr{
							{Expr: &nir.SignalExpr{Name: "signal2"}},
						},
						Out: nil,
					},
				},
			},
			expectedOpCount: 2,
			expectedOpKinds: []nir.OpKind{nir.OpEmit, nir.OpEmit},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &nir.Program{
				Signals: map[string]nir.Signal{},
				Rules: map[string]nir.Rule{
					"Test": tt.rule,
				},
			}

			dce := NewDeadCodeElimination()
			_, err := dce.Run(p)
			if err != nil {
				t.Fatalf("Run() returned error: %v", err)
			}

			rule := p.Rules["Test"]
			if len(rule.Ops) != tt.expectedOpCount {
				t.Errorf("Expected %d ops, got %d", tt.expectedOpCount, len(rule.Ops))
			}

			for i, expectedKind := range tt.expectedOpKinds {
				if i >= len(rule.Ops) {
					break
				}
				if rule.Ops[i].Kind != expectedKind {
					t.Errorf("Expected op %d to be %s, got %s", i, expectedKind, rule.Ops[i].Kind)
				}
			}
		})
	}
}

func TestDeadCodeElimination_ComplexExpressions(t *testing.T) {
	tests := []struct {
		name            string
		rule            nir.Rule
		expectedOpCount int
	}{
		{
			name: "tracks dependencies through IdxExpr",
			rule: nir.Rule{
				Symslots: 4,
				Outputs: map[string]nir.Output{
					"result": {Sym: 3},
				},
				Ops: []nir.Op{
					{
						// dead - result not used
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: core.V(42)}},
						},
						Out: core.IntPtr(0),
					},
					{
						// live - slot 1 creates an array
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{
								Expr: &nir.ArrExpr{
									Value: []nir.RawExpr{
										{Expr: &nir.ImmExpr{Value: core.V(100)}},
									},
								},
							},
						},
						Out: core.IntPtr(1),
					},
					{
						// live - reads from slot 1 through IdxExpr, writes to slot 2
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{
								Expr: &nir.IdxExpr{
									From:  nir.RawExpr{Expr: &nir.SymExpr{Index: 1}},
									Index: nir.RawExpr{Expr: &nir.ImmExpr{Value: core.V(0)}},
								},
							},
						},
						Out: core.IntPtr(2),
					},
					{
						// live - reads from slot 2, writes to output slot 3
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.SymExpr{Index: 2}},
						},
						Out: core.IntPtr(3),
					},
				},
			},
			expectedOpCount: 3,
		},
		{
			name: "tracks dependencies through ObjExpr",
			rule: nir.Rule{
				Symslots: 3,
				Outputs: map[string]nir.Output{
					"result": {Sym: 2},
				},
				Ops: []nir.Op{
					{
						// dead
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: core.V(42)}},
						},
						Out: core.IntPtr(0),
					},
					{
						// live - slot 1 is used in the object
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: core.V(100)}},
						},
						Out: core.IntPtr(1),
					},
					{
						// live - creates an object with a prop reading from slot 1, writes to output slot 2
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{
								Expr: &nir.ObjExpr{
									Value: map[string]nir.RawExpr{
										"a": {Expr: &nir.SymExpr{Index: 1}},
									},
								},
							},
						},
						Out: core.IntPtr(2),
					},
				},
			},
			expectedOpCount: 2,
		},
		{
			name: "tracks dependencies through ArrExpr",
			rule: nir.Rule{
				Symslots: 3,
				Outputs: map[string]nir.Output{
					"result": {Sym: 2},
				},
				Ops: []nir.Op{
					{
						// dead
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: core.V(42)}},
						},
						Out: core.IntPtr(0),
					},
					{
						// live - slot 1 is used in the array
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: core.V(100)}},
						},
						Out: core.IntPtr(1),
					},
					{
						// live - creates array reading from slot 1, writes to output slot 2
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{
								Expr: &nir.ArrExpr{
									Value: []nir.RawExpr{
										{Expr: &nir.SymExpr{Index: 1}},
									},
								},
							},
						},
						Out: core.IntPtr(2),
					},
				},
			},
			expectedOpCount: 2,
		},
		{
			name: "tracks dependencies through CallExpr",
			rule: nir.Rule{
				Symslots: 3,
				Outputs: map[string]nir.Output{
					"result": {Sym: 2},
				},
				Ops: []nir.Op{
					{
						// dead
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: core.V(42)}},
						},
						Out: core.IntPtr(0),
					},
					{
						// live - slot 1 is used in the array
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: core.V(100)}},
						},
						Out: core.IntPtr(1),
					},
					{
						// live - calls a function with slot 1 as argument, writes to output slot 2
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{
								Expr: &nir.CallExpr{
									Func: nir.CallFunc{}, // dummy
									Args: []nir.RawExpr{
										{Expr: &nir.SymExpr{Index: 1}},
									},
								},
							},
						},
						Out: core.IntPtr(2),
					},
				},
			},
			expectedOpCount: 2,
		},
		{
			name: "tracks dependencies through SelExpr",
			rule: nir.Rule{
				Symslots: 3,
				Outputs: map[string]nir.Output{
					"result": {Sym: 2},
				},
				Ops: []nir.Op{
					{
						// dead
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: core.V(42)}},
						},
						Out: core.IntPtr(0),
					},
					{
						// live - slot 1 is read in SelExpr
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: core.V(100)}},
						},
						Out: core.IntPtr(1),
					},
					{
						// live - reads from slot 1 through SelExpr, writes to output slot 2
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{
								Expr: &nir.SelExpr{
									Path: "field",
									From: nir.RawExpr{Expr: &nir.SymExpr{Index: 1}},
								},
							},
						},
						Out: core.IntPtr(2),
					},
				},
			},
			expectedOpCount: 2,
		},
		{
			name: "tracks dependencies through SignalExpr with When",
			rule: nir.Rule{
				Symslots: 3,
				Outputs: map[string]nir.Output{
					"result": {Sym: 2},
				},
				Ops: []nir.Op{
					{
						// dead
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: core.V(42)}},
						},
						Out: core.IntPtr(0),
					},
					{
						// live - slot 1 is read in SignalExpr When clause
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: core.V(true)}},
						},
						Out: core.IntPtr(1),
					},
					{
						// live - SignalExpr reads from slot 1 in When, writes to output slot 2
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{
								Expr: &nir.SignalExpr{
									Name: "test",
									When: &nir.RawExpr{Expr: &nir.SymExpr{Index: 1}},
								},
							},
						},
						Out: core.IntPtr(2),
					},
				},
			},
			expectedOpCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &nir.Program{
				Signals: map[string]nir.Signal{},
				Rules: map[string]nir.Rule{
					"Test": tt.rule,
				},
			}

			dce := NewDeadCodeElimination()
			_, err := dce.Run(p)
			if err != nil {
				t.Fatalf("Run() returned error: %v", err)
			}

			rule := p.Rules["Test"]
			if len(rule.Ops) != tt.expectedOpCount {
				t.Errorf("Expected %d ops, got %d", tt.expectedOpCount, len(rule.Ops))
			}
		})
	}
}

func TestDeadCodeElimination_MultipleRules(t *testing.T) {
	p := &nir.Program{
		Signals: map[string]nir.Signal{
			"usedInRule1": {Params: []nir.Param{{Name: "x"}}},
			"usedInRule2": {Params: []nir.Param{{Name: "y"}}},
			"unused":      {Params: []nir.Param{{Name: "z"}}},
		},
		Rules: map[string]nir.Rule{
			"Rule1": {
				Symslots: 1,
				Outputs: map[string]nir.Output{
					"output": {Sym: 0},
				},
				Ops: []nir.Op{
					{
						Kind: nir.OpEmit,
						Args: []nir.RawExpr{
							{Expr: &nir.SignalExpr{Name: "usedInRule1"}},
						},
					},
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: core.V(42)}},
						},
						Out: core.IntPtr(0),
					},
				},
			},
			"Rule2": {
				Symslots: 1,
				Outputs: map[string]nir.Output{
					"output": {Sym: 0},
				},
				Ops: []nir.Op{
					{
						Kind: nir.OpEmit,
						Args: []nir.RawExpr{
							{Expr: &nir.SignalExpr{Name: "usedInRule2"}},
						},
					},
					{
						Kind: nir.OpCopy,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: core.V(100)}},
						},
						Out: core.IntPtr(0),
					},
				},
			},
		},
	}

	dce := NewDeadCodeElimination()
	_, err := dce.Run(p)
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	// Check signals
	if len(p.Signals) != 2 {
		t.Errorf("Expected 2 signals, got %d", len(p.Signals))
	}
	if _, exists := p.Signals["usedInRule1"]; !exists {
		t.Error("Expected usedInRule1 signal to exist")
	}
	if _, exists := p.Signals["usedInRule2"]; !exists {
		t.Error("Expected usedInRule2 signal to exist")
	}
	if _, exists := p.Signals["unused"]; exists {
		t.Error("Expected unused signal to be removed")
	}

	// Check rules still exist
	if len(p.Rules) != 2 {
		t.Errorf("Expected 2 rules, got %d", len(p.Rules))
	}

	// Check op count in each rule
	for n, r := range p.Rules {
		if len(r.Ops) != 2 {
			t.Errorf("%s: Expected 2 ops, got %d", n, len(p.Rules))
		}
	}
}
