package nir

import (
	"fmt"
	"strings"

	version "nodora.org/nodora"
	"nodora.org/nodora/internal/ast"
	"nodora.org/nodora/pkg/core"
)

type exprResult struct {
	isImmediate bool
	symIndex    int
	expr        Expr
}

func immResult(val any) exprResult {
	return exprResult{isImmediate: true, expr: &ImmExpr{Value: core.V(val)}}
}

func arrResult(exprs []exprResult) exprResult {
	raw := make([]RawExpr, len(exprs))
	for i, e := range exprs {
		raw[i] = e.toRawExpr()
	}
	return exprResult{expr: &ArrExpr{Value: raw}}
}

func objResult(obj map[string]exprResult) exprResult {
	raw := make(map[string]RawExpr, len(obj))
	for i, e := range obj {
		raw[i] = e.toRawExpr()
	}
	return exprResult{expr: &ObjExpr{Value: raw}}
}

func symResult(index int) exprResult {
	return exprResult{symIndex: index, expr: &SymExpr{Index: index}}
}

func selResult(path string, from RawExpr, exprs []RawExpr) exprResult {
	return exprResult{expr: &SelExpr{Path: path, From: from, Exprs: exprs}}
}

func idxResult(e exprResult, idx exprResult) exprResult {
	return exprResult{expr: &IdxExpr{From: e.toRawExpr(), Index: idx.toRawExpr()}}
}

func callResult(name string, args ...exprResult) exprResult {
	rawArgs := make([]RawExpr, len(args))
	for i, a := range args {
		rawArgs[i] = a.toRawExpr()
	}
	return exprResult{expr: &CallExpr{Func: CallFunc{Name: name}, Args: rawArgs}}
}

func (r exprResult) toRawExpr() RawExpr {
	return RawExpr{Expr: r.expr}
}

func isImmBool(r exprResult, want bool) bool {
	imm, ok := r.expr.(*ImmExpr)
	if !ok {
		return false
	}
	b, ok := imm.Value.Raw.(bool)
	return ok && b == want
}

type Builder struct {
	symbols  map[string]int
	symIndex int
	outputs  map[string]bool
	consts   []*ast.Const
	ops      []Op
}

func NewBuilder() *Builder {
	return &Builder{
		symbols:  make(map[string]int),
		symIndex: 0,
		outputs:  make(map[string]bool),
		ops:      make([]Op, 0),
	}
}

func (b *Builder) Build(program *ast.Program) (*Ruleset, error) {
	meta := Metadata{"version": version.Version}
	result := &Ruleset{
		Metadata: meta,
		Signals:  make(map[string]Signal),
		Rules:    make(map[string]Rule),
	}

	for _, decl := range program.Decls {
		switch node := decl.(type) {
		case *ast.Const:
			b.consts = append(b.consts, node)
		case *ast.Signal:
			result.Signals[node.Name] = Signal{
				Params: b.buildParams(node.Params),
			}
		case *ast.Rule:
			rule, err := b.buildRule(node)
			if err != nil {
				return nil, fmt.Errorf("error converting rule %s: %w", node.Name, err)
			}
			result.Rules[node.Name] = rule
		}
	}

	return result, nil
}

func (b *Builder) reset() {
	b.symbols = make(map[string]int, 1)
	b.symIndex = 1 // 0 is reserved
	b.outputs = make(map[string]bool)
	b.ops = make([]Op, 0)
}

func (b *Builder) buildRule(rule *ast.Rule) (Rule, error) {
	b.reset()

	result := Rule{
		Outputs:  make(map[string]Output),
		Symslots: 1,
		Ops:      make([]Op, 0),
	}

	b.injectConsts()

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
		targetSym = b.materialize(result)
	}

	b.symbols[assign.Name] = targetSym
	return nil
}

// lowers the top-level consts visible to the current rule into its symbol space,
// in declaration order, at the top of the rule's op stream so references resolve
// like ordinary symbols
func (b *Builder) injectConsts() {
	for _, c := range b.consts {
		result := b.buildExpr(c.Value)

		var targetSym int
		if symExpr, ok := result.expr.(*SymExpr); ok && !result.isImmediate {
			targetSym = symExpr.Index
		} else {
			targetSym = b.materialize(result)
		}

		b.symbols[c.Name] = targetSym
	}
}

func (b *Builder) materialize(expr exprResult) int {
	newSym := b.allocSymbol()
	b.addOp(OpCopy, []RawExpr{expr.toRawExpr()}, &newSym)
	return newSym
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
		return immResult(e.Value)

	case *ast.StringLiteral:
		return immResult(e.Value)

	case *ast.ArrayLiteral:
		return b.buildArrayLiteral(e)

	case *ast.ObjectLiteral:
		return b.buildObjectLiteral(e)

	case *ast.BoolLiteral:
		return immResult(e.Value)

	case *ast.NullLiteral:
		return immResult(nil)

	case *ast.Identifier:
		if e.Name == "input" {
			return symResult(0)
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

		var from RawExpr
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
				from = b.buildExpr(e).toRawExpr()
			}
		}

		walk(expr)
		return selResult(strings.Join(parts[1:], "."), from, exprs)

	case *ast.IndexExpr:
		return idxResult(b.buildExpr(e.Expr), b.buildExpr(e.Index))

	case *ast.UnaryExpr:
		return b.buildUnaryExpr(e)

	case *ast.BinaryExpr:
		return b.buildBinaryExpr(e)

	case *ast.ConditionalExpr:
		return b.buildConditionalExpr(e)

	case *ast.MatchExpr:
		return b.buildMatchExpr(e)

	case *ast.CallExpr:
		return b.buildCallExpr(e)

	case *ast.LambdaExpr:
		return b.buildLambdaExpr(e)

	default:
		return exprResult{isImmediate: true, expr: nil}
	}
}

func (b *Builder) buildCallExpr(e *ast.CallExpr) exprResult {
	args := make([]RawExpr, len(e.Args))
	for i, arg := range e.Args {
		result := b.buildExpr(arg)
		args[i] = result.toRawExpr()
	}
	callFunc := CallFunc{
		Namespace: e.Namespace,
		Name:      e.Name,
	}
	return exprResult{expr: &CallExpr{
		Func: callFunc,
		Args: args,
	}}
}

func (b *Builder) buildUnaryExpr(expr *ast.UnaryExpr) exprResult {
	inner := b.buildExpr(expr.Expr)
	out := b.allocSymbol()

	switch expr.Op {
	case "-":
		b.addOp(OpSub, []RawExpr{{Expr: &ImmExpr{Value: core.V(0)}}, inner.toRawExpr()}, &out)
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

func (b *Builder) withBinding(name string, sym int, fn func()) {
	if name == "" {
		fn()
		return
	}
	prev, had := b.symbols[name]
	b.symbols[name] = sym
	fn()
	if had {
		b.symbols[name] = prev
	} else {
		delete(b.symbols, name)
	}
}

func (b *Builder) strictBool(cond exprResult) exprResult {
	defaulted := callResult("fallback", cond, immResult(false))
	matchSym := b.allocSymbol()
	b.addOp(OpEq, []RawExpr{defaulted.toRawExpr(), immResult(true).toRawExpr()}, &matchSym)
	return symResult(matchSym)
}

func (b *Builder) buildArmCondition(arm *ast.MatchArm, scrutSym int) exprResult {
	var cond exprResult
	hasLitCond := false
	if _, isIdent := arm.Pattern.(*ast.Identifier); !isIdent {
		patRes := b.buildExpr(arm.Pattern)
		eqSym := b.allocSymbol()
		b.addOp(OpEq, []RawExpr{symResult(scrutSym).toRawExpr(), patRes.toRawExpr()}, &eqSym)
		cond = symResult(eqSym)
		hasLitCond = true
	}

	if arm.Guard == nil {
		return cond
	}

	guard := b.buildExpr(arm.Guard)
	if hasLitCond {
		andSym := b.allocSymbol()
		b.addOp(OpAnd, []RawExpr{cond.toRawExpr(), guard.toRawExpr()}, &andSym)
		guard = symResult(andSym)
	}
	return b.strictBool(guard)
}

func (b *Builder) foldArm(arm *ast.MatchArm, scrutSym int, elseResult exprResult) exprResult {
	var cond, body exprResult
	b.withBinding(arm.BindingName(), scrutSym, func() {
		cond = b.buildArmCondition(arm, scrutSym)
		body = b.buildExpr(arm.Body)
	})

	if isImmBool(body, true) && isImmBool(elseResult, false) {
		return cond
	}

	selSym := b.allocSymbol()
	b.addOp(OpSelect, []RawExpr{cond.toRawExpr(), body.toRawExpr(), elseResult.toRawExpr()}, &selSym)
	return symResult(selSym)
}

func (b *Builder) buildMatchExpr(expr *ast.MatchExpr) exprResult {
	scrutSym := b.materialize(b.buildExpr(expr.Value))

	// use the last arm as the chain's terminal branch
	catchAllIdx := len(expr.Arms) - 1
	for i, arm := range expr.Arms {
		if arm.IsCatchAll() {
			catchAllIdx = i
			break
		}
	}

	catchArm := expr.Arms[catchAllIdx]
	var elseResult exprResult
	b.withBinding(catchArm.BindingName(), scrutSym, func() {
		elseResult = b.buildExpr(catchArm.Body)
	})

	for i := catchAllIdx - 1; i >= 0; i-- {
		elseResult = b.foldArm(expr.Arms[i], scrutSym, elseResult)
	}

	// gate on the scrutinee being defined: a defined value always resolves
	// through the arm chain (the catch-all is reachable), while an undefined
	// scrutinee yields undefined instead of falling into the catch-all
	gateSym := b.allocSymbol()
	scrutDefined := callResult("is_defined", symResult(scrutSym))
	b.addOp(OpSelect, []RawExpr{
		scrutDefined.toRawExpr(),
		elseResult.toRawExpr(),
		symResult(scrutSym).toRawExpr(),
	}, &gateSym)

	return symResult(gateSym)
}

func (b *Builder) buildConditionalExpr(expr *ast.ConditionalExpr) exprResult {
	cond := b.buildExpr(expr.Cond)
	thenBranch := b.buildExpr(expr.Then)
	elseBranch := b.buildExpr(expr.Else)

	out := b.allocSymbol()

	b.addOp(OpSelect, []RawExpr{cond.toRawExpr(), thenBranch.toRawExpr(), elseBranch.toRawExpr()}, &out)
	return symResult(out)
}

func (b *Builder) buildLambdaExpr(le *ast.LambdaExpr) exprResult {
	cb := NewBuilder()
	cb.symIndex = b.symIndex
	cb.symbols = b.symbols

	params := make([]LambdaParams, len(le.Params))
	for i, param := range le.Params {
		sym := cb.allocSymbol()
		cb.symbols[param.Name] = sym
		params[i] = LambdaParams{SymIndex: sym}
	}

	bodyExpr := cb.buildExpr(le.Body)
	_ = cb.materialize(bodyExpr)

	// restore symbol index in parent builder
	b.symIndex = cb.symIndex

	return exprResult{expr: &LambdaExpr{
		Params: params,
		Ops:    cb.ops,
	}}
}

func (b *Builder) buildArrayLiteral(lit *ast.ArrayLiteral) exprResult {
	elements := make([]exprResult, len(lit.Elements))
	for i, elem := range lit.Elements {
		result := b.buildExpr(elem)
		elements[i] = result
	}
	return arrResult(elements)
}

func (b *Builder) buildObjectLiteral(lit *ast.ObjectLiteral) exprResult {
	props := make(map[string]exprResult, len(lit.Properties))
	for _, prop := range lit.Properties {
		result := b.buildExpr(prop.Value)
		props[prop.Key] = result
	}
	return objResult(props)
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
