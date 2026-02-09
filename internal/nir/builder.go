package nir

import (
	"fmt"
	"strings"

	"nodora.org/nodora/internal/ast"
)

type exprResult struct {
	isImmediate bool
	symIndex    int
	expr        Expr
}

func immResult(val any) exprResult {
	return exprResult{isImmediate: true, expr: &ImmExpr{Value: val}}
}

func arrResult(exprs []exprResult) exprResult {
	raw := make([]RawExpr, len(exprs))
	for i, e := range exprs {
		raw[i] = e.toRawExpr()
	}
	return exprResult{expr: &ArrExpr{Value: raw}}
}

func symResult(index int) exprResult {
	return exprResult{symIndex: index, expr: &SymExpr{Index: index}}
}

func inputResult(path string, exprs []RawExpr) exprResult {
	return exprResult{expr: &InputExpr{Path: path, Exprs: exprs}}
}

func idxResult(e exprResult, idx exprResult) exprResult {
	return exprResult{expr: &IdxExpr{Value: []RawExpr{e.toRawExpr(), idx.toRawExpr()}}}
}

func (r exprResult) toRawExpr() RawExpr {
	return RawExpr{Expr: r.expr}
}

type Builder struct {
	symbols   map[string]int
	symIndex  int
	outputs   map[string]bool
	ops       []Op
	constants map[string]any
}

func NewBuilder() *Builder {
	return &Builder{
		symbols:   make(map[string]int),
		symIndex:  0,
		outputs:   make(map[string]bool),
		ops:       make([]Op, 0),
		constants: make(map[string]any),
	}
}

func (b *Builder) BuildFromAST(program *ast.Program) (Program, error) {
	result := Program{
		Signals: make(map[string]Signal),
		Rules:   make(map[string]Rule),
	}

	for _, decl := range program.Decls {
		switch node := decl.(type) {
		case *ast.Signal:
			result.Signals[node.Name] = Signal{
				Params: b.buildParams(node.Params),
			}
		case *ast.Rule:
			rule, err := b.buildRule(node)
			if err != nil {
				return Program{}, fmt.Errorf("error converting rule %s: %w", node.Name, err)
			}
			result.Rules[node.Name] = rule
		}
	}

	return result, nil
}

func (b *Builder) reset() {
	b.symbols = make(map[string]int)
	b.symIndex = 0
	b.outputs = make(map[string]bool)
	b.ops = make([]Op, 0)
	b.constants = make(map[string]any)
}

func (b *Builder) buildRule(rule *ast.Rule) (Rule, error) {
	b.reset()

	result := Rule{
		Outputs:  make(map[string]Output),
		Symbols:  make(SymbolMap),
		Symslots: 0,
		Ops:      make([]Op, 0),
	}

	for _, stmt := range rule.Statements {
		if err := b.buildStatement(stmt); err != nil {
			return Rule{}, err
		}
	}

	for outputName := range b.outputs {
		symIndex, exists := b.symbols[outputName]
		if !exists {
			return Rule{}, fmt.Errorf("output variable %s not found", outputName)
		}
		result.Outputs[outputName] = Output{Sym: symIndex}
	}

	result.Symbols = b.symbols
	result.Symslots = b.symIndex
	result.Ops = b.ops

	return result, nil
}

func (b *Builder) buildStatement(stmt ast.Statement) error {
	switch s := stmt.(type) {
	case *ast.Assignment:
		if s.IsOut {
			b.outputs[s.Name] = true
		}
		return b.buildAssignment(s)
	case *ast.EmitStatement:
		return b.buildEmitStatement(s)
	default:
		return fmt.Errorf("unsupported statement type: %T", s)
	}
}

func (b *Builder) addOp(kind OpKind, args []RawExpr, outSym *int) {
	b.ops = append(b.ops, Op{
		Kind: kind,
		Args: args,
		Out:  outSym,
	})
}

func (b *Builder) buildAssignment(assign *ast.Assignment) error {
	result := b.buildExpr(assign.Expr)

	var targetSym int
	if symExpr, ok := result.expr.(*SymExpr); ok && !result.isImmediate {
		// reuse the existing symbol
		targetSym = symExpr.Index
	} else {
		// allocate a new symbol for other expressions
		targetSym = b.allocSymbol()
		b.addOp(OpCopy, []RawExpr{result.toRawExpr()}, &targetSym)
	}

	b.symbols[assign.Name] = targetSym

	if result.isImmediate {
		b.constants[assign.Name] = result.expr.(*ImmExpr).Value
	}

	return nil
}

func (b *Builder) buildEmitStatement(emit *ast.EmitStatement) error {
	args := make([]RawExpr, len(emit.Args))
	for i, arg := range emit.Args {
		result := b.buildExpr(arg)
		args[i] = result.toRawExpr()
	}

	signalExpr := &SignalExpr{
		Name: emit.Signal,
		Args: args,
	}

	if emit.Condition != nil {
		condResult := b.buildExpr(emit.Condition)
		signalExpr.When = &RawExpr{Expr: condResult.expr}
	}

	b.addOp(OpEmit, []RawExpr{{Expr: signalExpr}}, nil)
	return nil
}

func (b *Builder) allocSymbol() int {
	sym := b.symIndex
	b.symIndex++
	return sym
}

func (b *Builder) buildExpr(expr ast.Expr) exprResult {
	if expr == nil {
		return exprResult{isImmediate: true, expr: nil}
	}

	switch e := expr.(type) {
	case *ast.NumberLiteral:
		switch e.Kind {
		case ast.FloatNumber:
			return immResult(e.Float)
		case ast.IntNumber:
			return immResult(e.Int)
		default:
			return immResult(e.Value)
		}

	case *ast.StringLiteral:
		return immResult(e.Value)

	case *ast.BoolLiteral:
		return immResult(e.Value)

	case *ast.Identifier:
		if val, isConst := b.constants[e.Name]; isConst {
			return immResult(val)
		}
		if symIndex, exists := b.symbols[e.Name]; exists {
			return symResult(symIndex)
		}
		// undefined identifier - allocate symbol
		sym := b.allocSymbol()
		b.symbols[e.Name] = sym
		return symResult(sym)

	case *ast.SelectorExpr:
		parts := []string{}
		exprs := []RawExpr{}

		var walk func(ast.Expr)
		walk = func(e ast.Expr) {
			switch n := e.(type) {
			case *ast.SelectorExpr:
				walk(n.Expr)
				parts = append(parts, n.Field)
			case *ast.IndexExpr:
				walk(n.Expr)
				parts = append(parts, "$")
				exprs = append(exprs, b.buildExpr(n.Index).toRawExpr())
			case *ast.Identifier:
				parts = append(parts, n.Name)
			}
		}

		walk(expr)
		inputless := parts[1:]
		return inputResult(strings.Join(inputless, "."), exprs)

	case *ast.IndexExpr:
		return idxResult(b.buildExpr(e.Expr), b.buildExpr(e.Index))

	case *ast.UnaryExpr:
		return b.buildUnaryExpr(e)

	case *ast.BinaryExpr:
		return b.buildBinaryExpr(e)

	case *ast.ConditionalExpr:
		return b.buildConditionalExpr(e)

	case *ast.ArrayLiteral:
		return b.buildArrayLiteral(e)

	default:
		return exprResult{isImmediate: true, expr: nil}
	}
}

func (b *Builder) buildUnaryExpr(expr *ast.UnaryExpr) exprResult {
	inner := b.buildExpr(expr.Expr)
	out := b.allocSymbol()

	switch expr.Op {
	case "-":
		b.addOp(OpSub, []RawExpr{{Expr: &ImmExpr{Value: 0}}, inner.toRawExpr()}, &out)
	case "!":
		b.addOp(OpNot, []RawExpr{inner.toRawExpr()}, &out)
	}

	return symResult(out)
}

func (b *Builder) buildBinaryExpr(expr *ast.BinaryExpr) exprResult {
	left := b.buildExpr(expr.Left)
	right := b.buildExpr(expr.Right)

	out := b.allocSymbol()

	opKind := b.mapBinaryOp(expr.Op)
	if opKind == "" {
		return symResult(out)
	}

	b.addOp(opKind, []RawExpr{left.toRawExpr(), right.toRawExpr()}, &out)
	return symResult(out)
}

func (b *Builder) buildConditionalExpr(expr *ast.ConditionalExpr) exprResult {
	cond := b.buildExpr(expr.Cond)
	thenBranch := b.buildExpr(expr.Then)
	elseBranch := b.buildExpr(expr.Else)

	out := b.allocSymbol()

	b.addOp(OpSelect, []RawExpr{cond.toRawExpr(), thenBranch.toRawExpr(), elseBranch.toRawExpr()}, &out)
	return symResult(out)
}

func (b *Builder) buildArrayLiteral(lit *ast.ArrayLiteral) exprResult {
	elements := make([]exprResult, len(lit.Elements))
	for i, elem := range lit.Elements {
		result := b.buildExpr(elem)
		elements[i] = result
	}
	return arrResult(elements)
}

func (b *Builder) mapBinaryOp(op string) OpKind {
	switch op {
	case "+":
		return OpAdd
	case "-":
		return OpSub
	case "*":
		return OpMul
	case "/":
		return OpDiv
	case "%":
		return OpMod
	case "&&":
		return OpAnd
	case "||":
		return OpOr
	case "<":
		return OpLt
	case "<=":
		return OpLte
	case ">":
		return OpGt
	case ">=":
		return OpGte
	case "==":
		return OpEq
	case "!=":
		return OpNeq
	case "in":
		return OpIn
	default:
		return ""
	}
}

func (b *Builder) buildParams(params []ast.Param) []Param {
	result := make([]Param, len(params))
	for i, p := range params {
		result[i] = Param{Name: p.Name}
	}
	return result
}
