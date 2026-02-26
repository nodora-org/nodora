package passes

import (
	"testing"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/nir"
)

type testCase struct {
	name     string
	opKind   nir.OpKind
	args     []core.Value
	expected core.Value
}

func TestConstantFolding(t *testing.T) {
	testCases := []testCase{
		{"add", nir.OpAdd, []core.Value{core.V(5), core.V(3)}, core.V(8)},
		{"sub", nir.OpSub, []core.Value{core.V(10), core.V(4)}, core.V(6)},
		{"mul", nir.OpMul, []core.Value{core.V(7), core.V(3)}, core.V(21)},
		{"div", nir.OpDiv, []core.Value{core.V(20), core.V(4)}, core.V(5)},
		{"mod", nir.OpMod, []core.Value{core.V(17), core.V(5)}, core.V(2)},
		{"add", nir.OpAdd, []core.Value{core.V(5.5), core.V(3.5)}, core.V(9.0)},
		{"mul", nir.OpMul, []core.Value{core.V(2.5), core.V(4.0)}, core.V(10.0)},
		{"concat strings", nir.OpAdd, []core.Value{core.V("hello"), core.V(" world")}, core.V("hello world")},
		{"and", nir.OpAnd, []core.Value{core.V(true), core.V(false)}, core.V(false)},
		{"or", nir.OpOr, []core.Value{core.V(true), core.V(false)}, core.V(true)},
		{"not true", nir.OpNot, []core.Value{core.V(true)}, core.V(false)},
		{"not false", nir.OpNot, []core.Value{core.V(false)}, core.V(true)},
		{"eq true", nir.OpEq, []core.Value{core.V(5), core.V(5)}, core.V(true)},
		{"eq false", nir.OpEq, []core.Value{core.V(5), core.V(3)}, core.V(false)},
		{"neq", nir.OpNeq, []core.Value{core.V(5), core.V(3)}, core.V(true)},
		{"lt", nir.OpLt, []core.Value{core.V(3), core.V(5)}, core.V(true)},
		{"gt", nir.OpGt, []core.Value{core.V(5), core.V(3)}, core.V(true)},
		{"lte equal", nir.OpLte, []core.Value{core.V(5), core.V(5)}, core.V(true)},
		{"gte", nir.OpGte, []core.Value{core.V(5), core.V(3)}, core.V(true)},
		{"select true", nir.OpSelect, []core.Value{core.V(true), core.V(10), core.V(20)}, core.V(10)},
		{"select false", nir.OpSelect, []core.Value{core.V(false), core.V(10), core.V(20)}, core.V(20)},
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
								Out:  core.IntPtr(outSym),
							},
						},
					},
				},
			}

			cf := NewConstantFolding()
			_, _ = cf.Run(p)

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

			if !core.SafeEquals(imm.Value, tt.expected) {
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
							{Expr: &nir.ImmExpr{Value: core.V(5)}},
						},
						Out: core.IntPtr(2),
					},
				},
			},
		},
	}

	cf := NewConstantFolding()
	_, _ = cf.Run(p)

	rule := p.Rules["Test"]
	op := rule.Ops[0]
	if op.Kind != nir.OpAdd {
		t.Errorf("Expected OpAdd (not folded), got %s", op.Kind)
	}
	if len(op.Args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(op.Args))
	}
}
