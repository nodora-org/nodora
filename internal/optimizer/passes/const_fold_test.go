package passes

import (
	"testing"

	"nodora.org/nodora/pkg/nir"
)

type testCase struct {
	name     string
	opKind   nir.OpKind
	args     []nir.Value
	expected nir.Value
}

func valuesEqual(a, b nir.Value) bool {
	return a == b
}

func TestConstantFolding(t *testing.T) {
	testCases := []testCase{
		{"add ints", nir.OpAdd, []nir.Value{5, 3}, int64(8)},
		{"sub ints", nir.OpSub, []nir.Value{10, 4}, int64(6)},
		{"mul ints", nir.OpMul, []nir.Value{7, 3}, int64(21)},
		{"div ints", nir.OpDiv, []nir.Value{20, 4}, int64(5)},
		{"mod ints", nir.OpMod, []nir.Value{17, 5}, int64(2)},
		{"add floats", nir.OpAdd, []nir.Value{5.5, 3.5}, 9.0},
		{"mul floats", nir.OpMul, []nir.Value{2.5, 4.0}, 10.0},
		{"concat strings", nir.OpAdd, []nir.Value{"hello", " world"}, "hello world"},
		{"and bools", nir.OpAnd, []nir.Value{true, false}, false},
		{"or bools", nir.OpOr, []nir.Value{true, false}, true},
		{"not true", nir.OpNot, []nir.Value{true}, false},
		{"not false", nir.OpNot, []nir.Value{false}, true},
		{"eq ints true", nir.OpEq, []nir.Value{5, 5}, true},
		{"eq ints false", nir.OpEq, []nir.Value{5, 3}, false},
		{"neq ints", nir.OpNeq, []nir.Value{5, 3}, true},
		{"lt ints", nir.OpLt, []nir.Value{3, 5}, true},
		{"gt ints", nir.OpGt, []nir.Value{5, 3}, true},
		{"lte ints equal", nir.OpLte, []nir.Value{5, 5}, true},
		{"gte ints", nir.OpGte, []nir.Value{5, 3}, true},
		{"select true", nir.OpSelect, []nir.Value{true, 10, 20}, 10},
		{"select false", nir.OpSelect, []nir.Value{false, 10, 20}, 20},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			outSym := 1
			args := []nir.RawExpr{}
			for _, v := range tt.args {
				args = append(args, nir.RawExpr{Expr: &nir.ImmExpr{Value: v}})
			}
			p := &nir.Program{
				Rules: map[string]nir.Rule{
					"Test": {
						Ops: []nir.Op{
							{
								Kind: tt.opKind,
								Args: args,
								Out:  nir.IntPtr(outSym),
							},
						},
					},
				},
			}

			cf := NewConstantFolding()
			cf.Run(p)

			rule := p.Rules["Test"]
			if len(rule.Ops) != 1 {
				t.Fatalf("Expected 1 op, got %d", len(rule.Ops))
			}

			op := rule.Ops[0]
			if op.Kind != nir.OpCopy {
				t.Errorf("Expected OpCopy after folding, got %s", op.Kind)
			}

			if len(op.Args) != 1 {
				t.Fatalf("Expected 1 arg, got %d", len(op.Args))
			}

			imm, ok := op.Args[0].Expr.(*nir.ImmExpr)
			if !ok {
				t.Fatalf("Expected ImmExpr, got %T", op.Args[0].Expr)
			}

			if !valuesEqual(imm.Value, tt.expected) {
				t.Errorf("Expected %v (%T), got %v (%T)", tt.expected, tt.expected, imm.Value, imm.Value)
			}
		})
	}
}

func TestConstantFolding_NoFold(t *testing.T) {
	p := &nir.Program{
		Rules: map[string]nir.Rule{
			"Test": {
				Ops: []nir.Op{
					{
						Kind: nir.OpAdd,
						Args: []nir.RawExpr{
							{Expr: &nir.SymExpr{Index: 1}}, // not constant
							{Expr: &nir.ImmExpr{Value: 5}},
						},
						Out: nir.IntPtr(2),
					},
				},
			},
		},
	}

	cf := NewConstantFolding()
	cf.Run(p)

	rule := p.Rules["Test"]
	op := rule.Ops[0]
	if op.Kind != nir.OpAdd {
		t.Errorf("Expected OpAdd (not folded), got %s", op.Kind)
	}
	if len(op.Args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(op.Args))
	}
}

func TestConstantFolding_DivisionByZero(t *testing.T) {
	p := &nir.Program{
		Rules: map[string]nir.Rule{
			"rule1": {
				Ops: []nir.Op{
					{
						// 10 / 0 - should not fold (division by zero)
						Kind: nir.OpDiv,
						Args: []nir.RawExpr{
							{Expr: &nir.ImmExpr{Value: 10}},
							{Expr: &nir.ImmExpr{Value: 0}},
						},
						Out: nir.IntPtr(0),
					},
				},
			},
		},
	}

	cf := NewConstantFolding()
	err := cf.Run(p)

	if err == nil {
		t.Errorf("Expected OpDiv to fail due to division by zero")
	}
}
