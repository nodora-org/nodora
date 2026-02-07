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

type Converter struct {
	symbols   map[string]int
	symIndex  int
	outputs   map[string]bool
	ops       []Op
	constants map[string]any
}

func NewConverter() *Converter {
	return &Converter{
		symbols:   make(map[string]int),
		symIndex:  0,
		outputs:   make(map[string]bool),
		ops:       make([]Op, 0),
		constants: make(map[string]any),
	}
}

func (c *Converter) ConvertFromAST(program *ast.Program) (Program, error) {
	result := Program{
		Signals: make(map[string]Signal),
		Rules:   make(map[string]Rule),
	}

	for _, decl := range program.Decls {
		switch node := decl.(type) {
		case *ast.Signal:
			result.Signals[node.Name] = Signal{
				Params: c.convertParams(node.Params),
			}
		case *ast.Rule:
			rule, err := c.convertRule(node)
			if err != nil {
				return Program{}, fmt.Errorf("error converting rule %s: %w", node.Name, err)
			}
			result.Rules[node.Name] = rule
		}
	}

	return result, nil
}

func (c *Converter) reset() {
	c.symbols = make(map[string]int)
	c.symIndex = 0
	c.outputs = make(map[string]bool)
	c.ops = make([]Op, 0)
	c.constants = make(map[string]any)
}

func (c *Converter) convertRule(rule *ast.Rule) (Rule, error) {
	c.reset()

	result := Rule{
		Outputs:  make(map[string]Output),
		Symbols:  make(SymbolMap),
		Symslots: 0,
		Ops:      make([]Op, 0),
	}

	for _, stmt := range rule.Statements {
		if err := c.convertStatement(stmt); err != nil {
			return Rule{}, err
		}
	}

	for outputName := range c.outputs {
		symIndex, exists := c.symbols[outputName]
		if !exists {
			return Rule{}, fmt.Errorf("output variable %s not found", outputName)
		}
		result.Outputs[outputName] = Output{Sym: symIndex}
	}

	result.Symbols = c.symbols
	result.Symslots = c.symIndex
	result.Ops = c.ops

	return result, nil
}

func (c *Converter) convertStatement(stmt ast.Statement) error {
	switch s := stmt.(type) {
	case *ast.Assignment:
		if s.IsOut {
			c.outputs[s.Name] = true
		}
		return c.convertAssignment(s)
	case *ast.EmitStatement:
		return c.convertEmitStatement(s)
	default:
		return fmt.Errorf("unsupported statement type: %T", s)
	}
}

func (c *Converter) addOp(kind OpKind, args []RawExpr, outSym *int) {
	c.ops = append(c.ops, Op{
		Kind: kind,
		Args: args,
		Out:  outSym,
	})
}

func (c *Converter) convertAssignment(assign *ast.Assignment) error {
	result := c.convertExpr(assign.Expr)

	var targetSym int
	if symExpr, ok := result.expr.(*SymExpr); ok && !result.isImmediate {
		// reuse the existing symbol
		targetSym = symExpr.Index
	} else {
		// allocate a new symbol for other expressions
		targetSym = c.allocSymbol()
		c.addOp(OpCopy, []RawExpr{result.toRawExpr()}, &targetSym)
	}

	c.symbols[assign.Name] = targetSym

	if result.isImmediate {
		c.constants[assign.Name] = result.expr.(*ImmExpr).Value
	}

	return nil
}

func (c *Converter) convertEmitStatement(emit *ast.EmitStatement) error {
	args := make([]RawExpr, len(emit.Args))
	for i, arg := range emit.Args {
		result := c.convertExpr(arg)
		args[i] = result.toRawExpr()
	}

	signalExpr := &SignalExpr{
		Name: emit.Signal,
		Args: args,
	}

	if emit.Condition != nil {
		condResult := c.convertExpr(emit.Condition)
		signalExpr.When = &RawExpr{Expr: condResult.expr}
	}

	c.addOp(OpEmit, []RawExpr{{Expr: signalExpr}}, nil)
	return nil
}

func (c *Converter) allocSymbol() int {
	sym := c.symIndex
	c.symIndex++
	return sym
}

func (c *Converter) convertExpr(expr ast.Expr) exprResult {
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
		if val, isConst := c.constants[e.Name]; isConst {
			return immResult(val)
		}
		if symIndex, exists := c.symbols[e.Name]; exists {
			return symResult(symIndex)
		}
		// undefined identifier - allocate symbol
		sym := c.allocSymbol()
		c.symbols[e.Name] = sym
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
				exprs = append(exprs, c.convertExpr(n.Index).toRawExpr())
			case *ast.Identifier:
				parts = append(parts, n.Name)
			}
		}

		walk(expr)
		inputless := parts[1:]
		return inputResult(strings.Join(inputless, "."), exprs)

	case *ast.IndexExpr:
		return idxResult(c.convertExpr(e.Expr), c.convertExpr(e.Index))

	case *ast.UnaryExpr:
		return c.convertUnaryExpr(e)

	case *ast.BinaryExpr:
		return c.convertBinaryExpr(e)

	case *ast.ConditionalExpr:
		return c.convertConditionalExpr(e)

	case *ast.ArrayLiteral:
		return c.convertArrayLiteral(e)

	default:
		return exprResult{isImmediate: true, expr: nil}
	}
}

func (c *Converter) convertUnaryExpr(expr *ast.UnaryExpr) exprResult {
	inner := c.convertExpr(expr.Expr)
	out := c.allocSymbol()

	switch expr.Op {
	case "-":
		c.addOp(OpSub, []RawExpr{{Expr: &ImmExpr{Value: 0}}, inner.toRawExpr()}, &out)
	case "!":
		c.addOp(OpNot, []RawExpr{inner.toRawExpr()}, &out)
	}

	return symResult(out)
}

func (c *Converter) convertBinaryExpr(expr *ast.BinaryExpr) exprResult {
	left := c.convertExpr(expr.Left)
	right := c.convertExpr(expr.Right)

	out := c.allocSymbol()

	opKind := c.mapBinaryOp(expr.Op)
	if opKind == "" {
		return symResult(out)
	}

	c.addOp(opKind, []RawExpr{left.toRawExpr(), right.toRawExpr()}, &out)
	return symResult(out)
}

func (c *Converter) convertConditionalExpr(expr *ast.ConditionalExpr) exprResult {
	cond := c.convertExpr(expr.Cond)
	thenBranch := c.convertExpr(expr.Then)
	elseBranch := c.convertExpr(expr.Else)

	out := c.allocSymbol()

	c.addOp(OpSelect, []RawExpr{cond.toRawExpr(), thenBranch.toRawExpr(), elseBranch.toRawExpr()}, &out)
	return symResult(out)
}

func (c *Converter) convertArrayLiteral(lit *ast.ArrayLiteral) exprResult {
	elements := make([]exprResult, len(lit.Elements))
	for i, elem := range lit.Elements {
		result := c.convertExpr(elem)
		elements[i] = result
	}
	return arrResult(elements)
}

func (c *Converter) mapBinaryOp(op string) OpKind {
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

func (c *Converter) convertParams(params []ast.Param) []Param {
	result := make([]Param, len(params))
	for i, p := range params {
		result[i] = Param{Name: p.Name}
	}
	return result
}
