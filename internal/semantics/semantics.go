package semantics

import (
	"fmt"
	"strconv"
	"strings"

	"nodora.org/nodora/internal/ast"
	"nodora.org/nodora/pkg/registry"
	"nodora.org/nodora/pkg/types"
)

const (
	symbolKindSignal   = "signal"
	symbolKindRule     = "rule"
	symbolKindVar      = "var"
	symbolKindConst    = "const"
	symbolKindReserved = "reserved"
)

type scope struct {
	symbols map[string]*symbol
	parent  *scope
}

type symbol struct {
	Name  string
	Type  types.Type
	Kind  string
	Value any
}

type SourceError struct {
	Span    ast.Span `json:"span"`
	Message string   `json:"message"`
	srcLine string
}

func (e *SourceError) Error() string {
	pos := e.Span.Start
	header := fmt.Sprintf("%d:%d: %s", pos.Line, pos.Col+1, e.Message)
	if e.srcLine == "" {
		return header
	}
	caret := strings.Repeat(" ", pos.Col) + "^"
	return fmt.Sprintf("%s\n%s\n%s", header, e.srcLine, caret)
}

type SemanticErrors struct {
	Errors []error `json:"errors"`
}

func (se *SemanticErrors) Add(err error) {
	se.Errors = append(se.Errors, err)
}

func (se *SemanticErrors) Count() int {
	return len(se.Errors)
}

func (se *SemanticErrors) Error() string {
	var sb strings.Builder

	for _, v := range se.Errors {
		sb.WriteString(v.Error())
		sb.WriteString("\n")
	}
	return sb.String()
}

type SemanticAnalyzer struct {
	scopes []*scope
	errors SemanticErrors
	lines  []string
}

func NewSemanticAnalyzer(src string) *SemanticAnalyzer {
	return &SemanticAnalyzer{
		scopes: []*scope{newScope(nil)}, // global scope
		errors: SemanticErrors{},
		lines:  strings.Split(src, "\n"),
	}
}

func newScope(parent *scope) *scope {
	input := &ast.ReservedObject{Name: "input"}
	input.Annotate(types.ObjectType)

	symbols := make(map[string]*symbol)
	symbols[input.Name] = &symbol{
		Name:  input.Name,
		Type:  types.ObjectType,
		Kind:  symbolKindReserved,
		Value: input,
	}
	return &scope{
		symbols: symbols,
		parent:  parent,
	}
}

func (sc *SemanticAnalyzer) pushScope() {
	newScope := newScope(sc.currentScope())
	sc.scopes = append(sc.scopes, newScope)
}

func (sc *SemanticAnalyzer) popScope() {
	if len(sc.scopes) > 1 {
		sc.scopes = sc.scopes[:len(sc.scopes)-1]
	}
}

func (sc *SemanticAnalyzer) currentScope() *scope {
	return sc.scopes[len(sc.scopes)-1]
}

func (sc *SemanticAnalyzer) isReservedSymbol(name string) bool {
	sym, _ := sc.lookupSymbol(name, ast.Span{})
	return sym != nil && sym.Kind == symbolKindReserved
}

func (sc *SemanticAnalyzer) declareSymbol(name string, typ types.Type, kind string, value any, span ast.Span) error {
	if sc.isReservedSymbol(name) {
		return sc.errorAt(span, "cannot declare symbol '%s': reserved", name)
	}
	scope := sc.currentScope()
	if sym, exists := scope.symbols[name]; exists {
		return sc.errorAt(span, "symbol '%s' already declared in this scope (%s)", name, sym.Kind)
	}
	scope.symbols[name] = &symbol{Name: name, Type: typ, Kind: kind, Value: value}
	return nil
}

func (sc *SemanticAnalyzer) lookupSymbol(name string, span ast.WithSpan) (*symbol, error) {
	for i := len(sc.scopes) - 1; i >= 0; i-- {
		if sym, exists := sc.scopes[i].symbols[name]; exists {
			return sym, nil
		}
	}
	return nil, sc.errorAt(span, "undefined symbol '%s'", name)
}

func (sc *SemanticAnalyzer) addError(err error) {
	sc.errors.Add(err)
}

// errorAt creates a SourceError with the source line and caret embedded.
func (sc *SemanticAnalyzer) errorAt(span ast.WithSpan, format string, args ...any) error {
	s := span.GetSpan()
	msg := fmt.Sprintf(format, args...)
	srcLine := ""
	if lineNum := s.Start.Line; lineNum >= 1 && lineNum <= len(sc.lines) {
		srcLine = sc.lines[lineNum-1]
	}
	return &SourceError{Span: s, Message: msg, srcLine: srcLine}
}

func (sc *SemanticAnalyzer) inferType(expr ast.Expr) types.Type {
	switch e := expr.(type) {
	case *ast.NumberLiteral:
		return types.NumberType
	case *ast.StringLiteral:
		return types.StringType
	case *ast.BoolLiteral:
		return types.BoolType
	case *ast.ArrayLiteral:
		return sc.inferArrayType(e)
	case *ast.ObjectLiteral:
		return types.ObjectType
	case *ast.Identifier:
		if sym, err := sc.lookupSymbol(e.Name, expr.GetSpan()); err == nil {
			return sym.Type
		}
		return types.UnknownType // symbol not found
	case *ast.BinaryExpr:
		if e.Type != nil {
			return e.Type
		}
		left := sc.inferType(e.Left)
		right := sc.inferType(e.Right)

		switch e.Op {
		case "+":
			if left.Equals(types.StringType) || right.Equals(types.StringType) {
				return types.StringType
			}
			return types.NumberType
		case "-", "*", "/", "%":
			return types.NumberType
		case "==", "!=", "<", "<=", ">", ">=", "&&", "||", "in":
			return types.BoolType
		default:
			return types.UnknownType
		}
	case *ast.UnaryExpr:
		if e.Type != nil {
			return e.Type
		}
		switch e.Op {
		case "!":
			return types.BoolType
		case "-":
			return types.NumberType
		default:
			return sc.inferType(e.Expr)
		}
	case *ast.ConditionalExpr:
		if e.Type != nil {
			return e.Type
		}
		// then and else should have the same type
		return sc.inferType(e.Then)
	case *ast.MatchExpr:
		if e.Type != nil {
			return e.Type
		}
		if len(e.Arms) > 0 {
			return sc.inferType(e.Arms[0].Body)
		}
		return types.UnknownType
	case *ast.CallExpr:
		if e.Type != nil {
			return e.Type
		}
		if fn, ok := registry.Global().Get(e.Namespace, e.Name); ok {
			return fn.ReturnType
		}
		return types.UnknownType
	case *ast.SelectorExpr:
		if e.Type != nil {
			return e.Type
		}
		return types.UnknownType
	case *ast.IndexExpr:
		if e.Type != nil {
			return e.Type
		}
		return types.UnknownType
	case *ast.LambdaExpr:
		if e.Type != nil {
			return e.Type
		}
		return sc.inferType(e.Body)
	default:
		return types.UnknownType
	}
}

func (sc *SemanticAnalyzer) inferArrayType(al *ast.ArrayLiteral) types.Type {
	if len(al.Elements) == 0 {
		return types.NewArrayType(types.AnyType)
	}

	firstType := sc.inferType(al.Elements[0])
	allSame := true

	for i := 1; i < len(al.Elements); i++ {
		elemType := sc.inferType(al.Elements[i])
		if !firstType.Equals(elemType) {
			allSame = false
			break
		}
	}

	if allSame {
		return types.NewArrayType(firstType)
	}

	// mixed array
	return types.NewArrayType(types.AnyType)
}

func (sc *SemanticAnalyzer) isType(actualType types.Type, expectedTypes ...types.Type) bool {
	if actualType.Equals(types.UnknownType) {
		return true
	}
	for _, expected := range expectedTypes {
		if expected.IsAssignableFrom(actualType) {
			return true
		}
	}
	return false
}

func (sc *SemanticAnalyzer) typesCompatible(a, b types.Type) bool {
	if a.Equals(types.UnknownType) || b.Equals(types.UnknownType) {
		return true
	}
	return a.IsAssignableFrom(b) || b.IsAssignableFrom(a)
}

func (sc *SemanticAnalyzer) Analyze(program *ast.Program) error {
	for _, decl := range program.Decls {
		switch d := decl.(type) {
		case *ast.Const:
			if err := d.Accept(sc); err != nil {
				sc.addError(err)
			}
		case *ast.Signal:
			if err := d.Accept(sc); err != nil {
				sc.addError(err)
			}
		case *ast.Rule:
			if err := decl.Accept(sc); err != nil {
				sc.addError(err)
			}
		}
	}

	if sc.errors.Count() > 0 {
		return &sc.errors
	}

	return nil
}

func (sc *SemanticAnalyzer) VisitProgram(p *ast.Program) error {
	return nil // handled in Analyze
}

func (sc *SemanticAnalyzer) VisitSignal(s *ast.Signal) error {
	if err := sc.declareSymbol(s.Name, nil, symbolKindSignal, s, s.Span); err != nil {
		sc.addError(err)
	}
	return nil
}

func (sc *SemanticAnalyzer) VisitConst(c *ast.Const) error {
	if err := c.Value.Accept(sc); err != nil {
		sc.addError(err)
	}

	if !sc.isConstantExpression(c.Value) {
		sc.addError(sc.errorAt(c.Value, "const '%s' must be a constant expression", c.Name))
	}

	ty := sc.inferType(c.Value)
	c.Annotate(ty)

	if err := sc.declareSymbol(c.Name, ty, symbolKindConst, c.Value, c.Span); err != nil {
		sc.addError(err)
	}

	return nil
}

// reports whether expr can be fully evaluated at compile time
// without access to any runtime value
func (sc *SemanticAnalyzer) isConstantExpression(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.NumberLiteral, *ast.StringLiteral, *ast.BoolLiteral:
		return true
	case *ast.ArrayLiteral:
		for _, elem := range e.Elements {
			if !sc.isConstantExpression(elem) {
				return false
			}
		}
		return true
	case *ast.ObjectLiteral:
		for _, prop := range e.Properties {
			if !sc.isConstantExpression(prop.Value) {
				return false
			}
		}
		return true
	case *ast.UnaryExpr:
		return sc.isConstantExpression(e.Expr)
	case *ast.BinaryExpr:
		return sc.isConstantExpression(e.Left) && sc.isConstantExpression(e.Right)
	case *ast.ConditionalExpr:
		return sc.isConstantExpression(e.Cond) &&
			sc.isConstantExpression(e.Then) &&
			sc.isConstantExpression(e.Else)
	case *ast.IndexExpr:
		return sc.isConstantExpression(e.Expr) && sc.isConstantExpression(e.Index)
	case *ast.SelectorExpr:
		return sc.isConstantExpression(e.Expr)
	case *ast.Identifier:
		sym, err := sc.lookupSymbol(e.Name, e)
		return err == nil && sym.Kind == symbolKindConst
	case *ast.CallExpr:
		fn, ok := registry.Global().Get(e.Namespace, e.Name)
		if !ok || !fn.Pure {
			return false
		}
		for _, arg := range e.Args {
			if !sc.isConstantExpression(arg) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (sc *SemanticAnalyzer) VisitRule(r *ast.Rule) error {
	if err := sc.declareSymbol(r.Name, nil, symbolKindRule, r, r.Span); err != nil {
		sc.addError(err)
	}

	sc.pushScope()
	for _, stmt := range r.Statements {
		if err := stmt.Accept(sc); err != nil {
			sc.addError(err)
		}
	}

	sc.popScope()
	return nil
}

func (sc *SemanticAnalyzer) VisitAssignment(a *ast.Assignment) error {
	if err := a.Expr.Accept(sc); err != nil {
		sc.addError(err)
	}

	ty := sc.inferType(a.Expr)

	if ty.Kind() == types.LambdaKind && a.IsOut {
		sc.addError(sc.errorAt(a.Span, "'out' requires a non-function value, got %v", ty.String()))
	}

	// inferType returns UnknownType on failure, which is compatible with all type checks
	a.Annotate(ty)

	if err := sc.declareSymbol(a.Name, a.Type, symbolKindVar, a.Expr, a.Span); err != nil {
		sc.addError(err)
	}

	return nil
}

func (sc *SemanticAnalyzer) VisitEmitStatement(es *ast.EmitStatement) error {
	for _, arg := range es.Args {
		if err := arg.Accept(sc); err != nil {
			sc.addError(err)
		}
	}

	if es.Condition != nil {
		if err := es.Condition.Accept(sc); err != nil {
			sc.addError(err)
		}
	}

	signalSym, err := sc.lookupSymbol(es.Signal, es.GetSpan())
	if err != nil {
		sc.addError(sc.errorAt(es.Span, "undefined signal '%s'", es.Signal))
		return nil
	}

	if signalSym.Kind != "signal" {
		sc.addError(sc.errorAt(es.Span, "'%s' is not a signal", es.Signal))
		return nil
	}

	signal := signalSym.Value.(*ast.Signal)
	if len(es.Args) != len(signal.Params) {
		sc.addError(sc.errorAt(es.Span, "signal '%s' expects %d arguments, got %d", es.Signal, len(signal.Params), len(es.Args)))
		return nil
	}

	for _, arg := range es.Args {
		if ident, ok := arg.(*ast.Identifier); ok {
			if _, err := sc.lookupSymbol(ident.Name, ident.Span); err != nil {
				sc.addError(err)
			}
		}
	}

	return nil
}

func (sc *SemanticAnalyzer) VisitBinaryExpr(be *ast.BinaryExpr) error {
	if err := be.Left.Accept(sc); err != nil {
		sc.addError(err)
	}
	if err := be.Right.Accept(sc); err != nil {
		sc.addError(err)
	}

	leftType := sc.inferType(be.Left)
	rightType := sc.inferType(be.Right)

	switch be.Op {
	case "+":
		valid := sc.isType(leftType, types.NumberType, types.StringType) &&
			sc.isType(rightType, types.NumberType, types.StringType) &&
			sc.typesCompatible(leftType, rightType)

		if !valid {
			sc.addError(sc.errorAt(be, "operator '%s' cannot be applied to '%s' and '%s'", be.Op, leftType.String(), rightType.String()))
		}

		if !leftType.Equals(types.UnknownType) {
			be.Annotate(leftType)
		} else {
			be.Annotate(rightType)
		}
	case "-", "*", "/", "%":
		lok := sc.isType(leftType, types.NumberType)
		rok := sc.isType(rightType, types.NumberType)
		if !(lok && rok) {
			sc.addError(sc.errorAt(be, "operator '%s' cannot be applied to '%s' and '%s'", be.Op, leftType.String(), rightType.String()))
		}
		be.Annotate(types.NumberType)
	case "==", "!=":
		if !sc.typesCompatible(leftType, rightType) {
			sc.addError(sc.errorAt(be, "cannot compare '%s' with '%s' using '%s'", leftType.String(), rightType.String(), be.Op))
		}
		be.Annotate(types.BoolType)
	case "<", "<=", ">", ">=":
		lok := sc.isType(leftType, types.NumberType)
		rok := sc.isType(rightType, types.NumberType)
		if !(lok && rok) {
			sc.addError(sc.errorAt(be, "operator '%s' cannot be applied to '%s' and '%s'", be.Op, leftType.String(), rightType.String()))
		}
		be.Annotate(types.BoolType)
	case "&&", "||":
		lok := sc.isType(leftType, types.BoolType)
		rok := sc.isType(rightType, types.BoolType)
		if !(lok && rok) {
			sc.addError(sc.errorAt(be, "operator '%s' cannot be applied to '%s' and '%s'", be.Op, leftType.String(), rightType.String()))
		}
		be.Annotate(types.BoolType)
	case "in":
		if !sc.isType(rightType, types.NewArrayType(types.AnyType)) {
			sc.addError(sc.errorAt(be, "operator '%s' cannot be applied to type '%s'", be.Op, rightType.String()))
		}
		be.Annotate(types.BoolType)
	default:
		sc.addError(sc.errorAt(be, "invalid binary operator '%s'", be.Op))
	}

	return nil
}

func (sc *SemanticAnalyzer) VisitUnaryExpr(ue *ast.UnaryExpr) error {
	if err := ue.Expr.Accept(sc); err != nil {
		sc.addError(err)
	}

	ty := sc.inferType(ue.Expr)
	switch ue.Op {
	case "!":
		if !sc.isType(ty, types.BoolType) {
			sc.addError(sc.errorAt(ue, "operand of '!' must be a '%s', got %s", types.BoolType.String(), ty.String()))
		}
		ue.Annotate(types.BoolType)
	case "-":
		if !sc.isType(ty, types.NumberType) {
			sc.addError(sc.errorAt(ue, "operand of '-' must be a '%s', got %s", types.NumberType.String(), ty.String()))
		}
		ue.Annotate(types.NumberType)
	default:
		sc.addError(sc.errorAt(ue, "invalid unary operator '%s'", ue.Op))
	}

	return nil
}

func (sc *SemanticAnalyzer) VisitIdentifier(id *ast.Identifier) error {
	if _, err := sc.lookupSymbol(id.Name, id.Span); err != nil {
		sc.addError(err)
	}
	return nil
}

func (sc *SemanticAnalyzer) VisitNumberLiteral(nl *ast.NumberLiteral) error {
	if strings.Contains(nl.Raw, ".") {
		val, err := strconv.ParseFloat(nl.Raw, 64)
		if err != nil {
			return err
		}
		nl.Value = val
		return nil
	}

	val, err := strconv.ParseInt(nl.Raw, 10, 64)
	if err != nil {
		return err
	}

	nl.Value = float64(val)
	return nil
}

func (sc *SemanticAnalyzer) VisitStringLiteral(sl *ast.StringLiteral) error {
	return nil
}

func (sc *SemanticAnalyzer) VisitBoolLiteral(bl *ast.BoolLiteral) error {
	return nil
}

func (sc *SemanticAnalyzer) isMatchExhaustive(scrutType types.Type, arms []*ast.MatchArm) bool {
	for _, arm := range arms {
		if arm.IsCatchAll() {
			return true
		}
	}
	// Bool exhaustiveness: both true and false literal arms with no guards
	// cover every bool value. Only applied when the scrutinee is statically
	// known to be bool — an unknown type might be anything at runtime.
	if scrutType != nil && scrutType.Equals(types.BoolType) {
		hasTrue, hasFalse := false, false
		for _, arm := range arms {
			if arm.Guard != nil {
				continue
			}
			if bl, ok := arm.Pattern.(*ast.BoolLiteral); ok {
				if bl.Value {
					hasTrue = true
				} else {
					hasFalse = true
				}
			}
		}
		if hasTrue && hasFalse {
			return true
		}
	}
	return false
}

func (sc *SemanticAnalyzer) VisitMatchExpr(m *ast.MatchExpr) error {
	if err := m.Value.Accept(sc); err != nil {
		sc.addError(err)
	}
	scrutType := sc.inferType(m.Value)

	if !sc.isMatchExhaustive(scrutType, m.Arms) {
		sc.addError(sc.errorAt(m, "match expression is not exhaustive: add a wildcard '_' or bare identifier arm"))
	}

	var firstBodyType types.Type
	for i, arm := range m.Arms {
		sc.pushScope()

		switch p := arm.Pattern.(type) {
		case *ast.Identifier:
			if p.Name != "_" {
				if err := sc.declareSymbol(p.Name, scrutType, symbolKindVar, m.Value, p.Span); err != nil {
					sc.addError(err)
				}
			}
		default:
			if err := arm.Pattern.Accept(sc); err != nil {
				sc.addError(err)
			}
			patType := sc.inferType(arm.Pattern)
			if !sc.typesCompatible(scrutType, patType) {
				sc.addError(sc.errorAt(arm.Pattern, "match pattern type '%s' is incompatible with scrutinee type '%s'", patType.String(), scrutType.String()))
			}
		}

		if arm.Guard != nil {
			if err := arm.Guard.Accept(sc); err != nil {
				sc.addError(err)
			}
			guardType := sc.inferType(arm.Guard)
			if !sc.isType(guardType, types.BoolType) {
				sc.addError(sc.errorAt(arm.Guard, "match guard must be of type bool, got '%s'", guardType.String()))
			}
		}

		if err := arm.Body.Accept(sc); err != nil {
			sc.addError(err)
		}
		bodyType := sc.inferType(arm.Body)
		if i == 0 {
			firstBodyType = bodyType
		} else if !sc.typesCompatible(firstBodyType, bodyType) {
			sc.addError(sc.errorAt(arm.Body, "incompatible match arm body types: '%s' and '%s'", firstBodyType.String(), bodyType.String()))
		}

		sc.popScope()
	}

	if firstBodyType != nil {
		m.Annotate(firstBodyType)
	} else {
		m.Annotate(types.UnknownType)
	}
	return nil
}

func (sc *SemanticAnalyzer) VisitConditionalExpr(ce *ast.ConditionalExpr) error {
	if err := ce.Cond.Accept(sc); err != nil {
		sc.addError(err)
	}
	condType := sc.inferType(ce.Cond)
	if !sc.isType(condType, types.BoolType) {
		sc.addError(sc.errorAt(ce, "condition must be of type bool, got '%s'", condType.String()))
	}

	if err := ce.Then.Accept(sc); err != nil {
		sc.addError(err)
	}
	if err := ce.Else.Accept(sc); err != nil {
		sc.addError(err)
	}

	thenType := sc.inferType(ce.Then)
	elseType := sc.inferType(ce.Else)

	if !sc.typesCompatible(thenType, elseType) {
		sc.addError(sc.errorAt(ce,
			"incompatible types in conditional expression, got '%s' and '%s'",
			thenType.String(),
			elseType.String(),
		))
	}

	ce.Annotate(thenType)
	return nil
}

func (sc *SemanticAnalyzer) VisitSelectorExpr(se *ast.SelectorExpr) error {
	if err := se.Expr.Accept(sc); err != nil {
		sc.addError(err)
		return nil
	}

	baseType := sc.inferType(se.Expr)
	if baseType.Equals(types.UnknownType) {
		se.Annotate(types.UnknownType)
		return nil
	}

	if !sc.isType(baseType, types.ObjectType) {
		sc.addError(sc.errorAt(se, "cannot access property '%s' on type %s", se.Field, baseType.String()))
		se.Annotate(types.UnknownType)
		return nil
	}

	propertyType := sc.getPropertyType(se.Expr, se.Field)
	se.Annotate(propertyType)
	return nil
}

func (sc *SemanticAnalyzer) getPropertyType(expr ast.Expr, field string) types.Type {
	resolved, path := sc.resolveExpr(expr)

	if obj, ok := resolved.(*ast.ObjectLiteral); ok {
		for _, prop := range obj.Properties {
			if prop.Key == field {
				return sc.inferType(prop.Value)
			}
		}
		sc.addError(sc.errorAt(expr, "unknown property '%s' on object '%s'", field, path))
		return types.UnknownType
	}

	// dynamic object (e.g. result of a function call)
	// can't statically determine property types
	return types.UnknownType
}

func (sc *SemanticAnalyzer) resolveExpr(expr ast.Expr) (ast.Expr, string) {
	switch e := expr.(type) {
	case *ast.Identifier:
		sym, err := sc.lookupSymbol(e.Name, e)
		if err != nil {
			return e, e.Name
		}
		resolved, _ := sc.resolveExpr(sym.Value.(ast.Expr))
		return resolved, e.Name

	case *ast.SelectorExpr:
		base, baseName := sc.resolveExpr(e.Expr)
		path := baseName + "." + e.Field

		if obj, ok := base.(*ast.ObjectLiteral); ok {
			for _, prop := range obj.Properties {
				if prop.Key == e.Field {
					resolved, _ := sc.resolveExpr(prop.Value)
					return resolved, path
				}
			}
		}
		return e, path

	case *ast.IndexExpr:
		base, baseName := sc.resolveExpr(e.Expr)
		index, _ := sc.resolveExpr(e.Index)

		// build the path representation
		indexStr := ""
		switch idx := index.(type) {
		case *ast.StringLiteral:
			indexStr = fmt.Sprintf("[%q]", idx.Value)
		case *ast.NumberLiteral:
			indexStr = fmt.Sprintf("[%v]", int(idx.Value))
		case *ast.Identifier:
			indexStr = fmt.Sprintf("[%s]", idx.Name)
		default:
			indexStr = "[?]"
		}
		path := baseName + indexStr

		switch baseResolved := base.(type) {
		case *ast.ArrayLiteral:
			// handle numeric index on array
			if numIdx, ok := index.(*ast.NumberLiteral); ok {
				if numIdx.IsInt() {
					idx := int(numIdx.Value)
					if idx >= 0 && idx < len(baseResolved.Elements) {
						resolved, _ := sc.resolveExpr(baseResolved.Elements[idx])
						return resolved, path
					}
				}
			}
			return e, path

		case *ast.ObjectLiteral:
			// handle string index on object (like obj["key"])
			var keyName string
			switch idx := index.(type) {
			case *ast.StringLiteral:
				keyName = idx.Value
			case *ast.Identifier:
				// could resolve the identifier to get its value
				keyName = idx.Name
			default:
				return e, path
			}

			for _, prop := range baseResolved.Properties {
				if prop.Key == keyName {
					resolved, _ := sc.resolveExpr(prop.Value)
					return resolved, path
				}
			}
			return e, path

		default:
			return e, path
		}

	default:
		return e, ""
	}
}

func (sc *SemanticAnalyzer) VisitIndexExpr(ie *ast.IndexExpr) error {
	if err := ie.Expr.Accept(sc); err != nil {
		sc.addError(err)
	}
	if err := ie.Index.Accept(sc); err != nil {
		sc.addError(err)
	}

	resolvedIdx, _ := sc.resolveExpr(ie.Index)
	baseType := sc.inferType(ie.Expr)
	indexType := sc.inferType(resolvedIdx)

	if baseType.Equals(types.UnknownType) || indexType.Equals(types.UnknownType) {
		ie.Annotate(types.UnknownType)
		return nil
	}

	if sc.isType(baseType, types.NewArrayType(types.AnyType)) {
		if !sc.isType(indexType, types.NumberType) {
			sc.addError(sc.errorAt(ie,
				"cannot index with type '%s' for '%s' (expected '%s')",
				indexType.String(), baseType.String(), types.NumberType.String(),
			))
			ie.Annotate(types.UnknownType)
			return nil
		}

		if num, ok := resolvedIdx.(*ast.NumberLiteral); ok && !num.IsInt() {
			sc.addError(sc.errorAt(ie.Index, "array index must be an integer"))
			ie.Annotate(types.UnknownType)
			return nil
		}

		if idx, ok := sc.getConstantIndex(resolvedIdx); ok {
			if resolved, _ := sc.resolveExpr(ie.Expr); resolved != nil {
				if arrLit, ok := resolved.(*ast.ArrayLiteral); ok {
					if idx < 0 || idx >= len(arrLit.Elements) {
						sc.addError(sc.errorAt(ie, "array index %d out of bounds (length %d)", idx, len(arrLit.Elements)))
						ie.Annotate(types.UnknownType)
						return nil
					}
					ie.Annotate(sc.inferType(arrLit.Elements[idx]))
					return nil
				}
			}
		}

		if arrayType, ok := baseType.(*types.ArrayType); ok {
			ie.Annotate(arrayType.Element)
		} else {
			ie.Annotate(types.AnyType)
		}
		return nil
	}

	if sc.isType(baseType, types.ObjectType) {
		if !sc.isType(indexType, types.StringType) {
			sc.addError(sc.errorAt(ie,
				"cannot index with type '%s' for '%s' (expected '%s')",
				indexType.String(), baseType.String(), types.StringType.String(),
			))
			ie.Annotate(types.UnknownType)
			return nil
		}

		if strLit, ok := resolvedIdx.(*ast.StringLiteral); ok {
			ie.Annotate(sc.getPropertyType(ie.Expr, strLit.Value))
		} else {
			ie.Annotate(types.AnyType)
		}
		return nil
	}

	sc.addError(sc.errorAt(ie.Index, "cannot index type '%s'", baseType.String()))
	ie.Annotate(types.UnknownType)
	return nil
}

func (sc *SemanticAnalyzer) getConstantIndex(expr ast.Expr) (int, bool) {
	switch e := expr.(type) {
	case *ast.NumberLiteral:
		return int(e.Value), true
	case *ast.UnaryExpr:
		if e.Op == "-" {
			val, ok := sc.getConstantIndex(e.Expr)
			if ok {
				return -val, true
			}
		}
	}
	return 0, false
}

func (sc *SemanticAnalyzer) VisitArrayLiteral(al *ast.ArrayLiteral) error {
	for _, elem := range al.Elements {
		if err := elem.Accept(sc); err != nil {
			sc.addError(err)
		}
	}
	al.Annotate(sc.inferArrayType(al))
	return nil
}

func (sc *SemanticAnalyzer) VisitObjectLiteral(al *ast.ObjectLiteral) error {
	seen := make(map[string]bool)
	for _, prop := range al.Properties {
		if seen[prop.Key] {
			sc.addError(sc.errorAt(prop, "duplicate object property key '%s'", prop.Key))
		} else {
			seen[prop.Key] = true
		}
		if err := prop.Value.Accept(sc); err != nil {
			sc.addError(err)
		}
		prop.Annotate(sc.inferType(prop.Value))
	}
	al.Annotate(types.ObjectType)
	return nil
}

func (sc *SemanticAnalyzer) VisitParam(p *ast.Param) error {
	return nil
}

func (sc *SemanticAnalyzer) VisitLambdaExpr(le *ast.LambdaExpr) error {
	sc.pushScope()
	for _, param := range le.Params {
		if err := sc.declareSymbol(param.Name, types.UnknownType, symbolKindVar, nil, le.Span); err != nil {
			sc.addError(err)
		}
	}
	if err := le.Body.Accept(sc); err != nil {
		sc.addError(err)
	}

	retType := sc.inferType(le.Body)
	sc.popScope()

	params := make([]types.Type, len(le.Params))
	for i := range le.Params {
		params[i] = types.AnyType
	}

	le.Annotate(types.NewLambdaType(params, retType))
	return nil
}

func (sc *SemanticAnalyzer) VisitCallExpr(ce *ast.CallExpr) error {
	if !registry.Global().Exists(ce.Namespace, ce.Name) {
		sc.addError(sc.errorAt(ce, "undefined function '%s'", ce.FullPath()))

		// still visit args to surface any errors within them
		for _, arg := range ce.Args {
			if err := arg.Accept(sc); err != nil {
				sc.addError(err)
			}
		}
		return nil
	}

	fn, _ := registry.Global().Get(ce.Namespace, ce.Name)
	reqArgCount := fn.RequiredArgCount()

	if len(ce.Args) < reqArgCount {
		sc.addError(sc.errorAt(ce, "function '%s' expects %d arguments, got %d",
			fn.FullPath(),
			reqArgCount,
			len(ce.Args),
		))
	}

	for i, arg := range ce.Args {
		if err := arg.Accept(sc); err != nil {
			sc.addError(err)
		}

		// only type-check args that have a corresponding parameter
		if i < len(fn.Args) {
			argType := sc.inferType(arg)
			fnArg := fn.Args[i]

			if !sc.isType(argType, fnArg.Type) {
				sc.addError(sc.errorAt(
					arg,
					"%v requires '%v' to be of type %v, got %v",
					fn.CompactSignature(),
					fnArg.Name,
					fnArg.Type,
					argType,
				))
			}
		}
	}

	ce.Annotate(sc.narrowReturnType(fn, ce.Args))
	return nil
}

// narrowReturnType attempts to narrow a union return type based on actual argument types
func (sc *SemanticAnalyzer) narrowReturnType(fn *types.Func, args []ast.Expr) types.Type {
	retUnion, ok := fn.ReturnType.(*types.UnionType)
	if !ok {
		return fn.ReturnType
	}

	for i, arg := range args {
		if i >= len(fn.Args) {
			break
		}
		paramUnion, ok := fn.Args[i].Type.(*types.UnionType)
		if !ok || len(paramUnion.Types) != len(retUnion.Types) {
			continue
		}

		argType := sc.inferType(arg)
		matchIdx := -1
		for j, member := range paramUnion.Types {
			if member.IsAssignableFrom(argType) {
				if matchIdx != -1 {
					// can't narrow because it matches multiple members
					matchIdx = -1
					break
				}
				matchIdx = j
			}
		}

		if matchIdx >= 0 {
			return retUnion.Types[matchIdx]
		}
	}

	return fn.ReturnType
}

func (sc *SemanticAnalyzer) VisitReservedObject(ro *ast.ReservedObject) error {
	return nil
}
