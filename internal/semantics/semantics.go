package semantics

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"nodora.org/nodora/internal/ast"
	"nodora.org/nodora/pkg/registry"
)

type scope struct {
	symbols map[string]*symbol
	parent  *scope
}

type symbol struct {
	Name  string
	Type  string
	Kind  string
	Value any
}

type SemanticErrors struct {
	Errors []error
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
}

func NewSemanticAnalyzer() *SemanticAnalyzer {
	return &SemanticAnalyzer{
		scopes: []*scope{newScope(nil)}, // global scope
		errors: SemanticErrors{},
	}
}

func newScope(parent *scope) *scope {
	symbols := make(map[string]*symbol)
	symbols["input"] = &symbol{
		Name:  "input",
		Type:  "object",
		Kind:  "reserved",
		Value: make(map[string]any),
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

func (sc *SemanticAnalyzer) isReservedName(name string) bool {
	return name == "input"
}

func (sc *SemanticAnalyzer) declareSymbol(name, typ, kind string, value any, span ast.Span) error {
	if sc.isReservedName(name) {
		return errorAt(span, "symbol '%s' is a reserved name", name)
	}
	scope := sc.currentScope()
	if _, exists := scope.symbols[name]; exists {
		return errorAt(span, "symbol '%s' already declared in this scope", name)
	}
	scope.symbols[name] = &symbol{Name: name, Type: typ, Kind: kind, Value: value}
	return nil
}

func (sc *SemanticAnalyzer) lookupSymbol(name string, span ast.Span) (*symbol, error) {
	for i := len(sc.scopes) - 1; i >= 0; i-- {
		if sym, exists := sc.scopes[i].symbols[name]; exists {
			return sym, nil
		}
	}
	return nil, errorAt(span, "undefined symbol '%s'", name)
}

func (sc *SemanticAnalyzer) addError(err error) {
	sc.errors.Add(err)
}

func errorAt(span ast.WithSpan, format string, args ...any) error {
	pos := span.GetSpan().Start
	return fmt.Errorf("%d:%d: %s", pos.Line, pos.Col+1, fmt.Sprintf(format, args...))
}

func (sc *SemanticAnalyzer) inferType(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.NumberLiteral:
		return "number"
	case *ast.StringLiteral:
		return "string"
	case *ast.BoolLiteral:
		return "bool"
	case *ast.Identifier:
		if sym, err := sc.lookupSymbol(e.Name, expr.GetSpan()); err == nil {
			return sym.Type
		}
		return "unknown" // symbol not found
	case *ast.BinaryExpr:
		if e.Type != "" {
			return e.Type
		}
		left := sc.inferType(e.Left)
		right := sc.inferType(e.Right)
		switch e.Op {
		case "+":
			if left == "string" || right == "string" {
				return "string"
			}
			return "number"
		case "-", "*", "/", "%":
			return "number"
		case "==", "!=", "<", "<=", ">", ">=", "&&", "||", "in":
			return "bool"
		default:
			return "unknown"
		}
	case *ast.UnaryExpr:
		if e.Type != "" {
			return e.Type
		}
		switch e.Op {
		case "!":
			return "bool"
		case "-":
			return "number"
		default:
			return sc.inferType(e.Expr)
		}
	case *ast.ConditionalExpr:
		if e.Type != "" {
			return e.Type
		}
		// assuming then and else have same type
		return sc.inferType(e.Then)
	case *ast.ArrayLiteral:
		return "array"
	case *ast.CallExpr:
		if fn, ok := registry.Get(e.Namespace, e.Name); ok {
			return fn.ReturnType
		}
		return "unknown"
	default:
		return "unknown"
	}
}

func (sc *SemanticAnalyzer) requireType(actualType string, expectedTypes ...string) bool {
	if actualType == "unknown" {
		return true
	}
	return slices.Contains(expectedTypes, actualType)
}

func (sc *SemanticAnalyzer) typesCompatible(a, b string) bool {
	if a == "unknown" || b == "unknown" {
		return true
	}
	return a == b
}

func (sc *SemanticAnalyzer) Analyze(program *ast.Program) error {
	for _, decl := range program.Decls {
		switch d := decl.(type) {
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
	if err := sc.declareSymbol(s.Name, "", "signal", s, s.Span); err != nil {
		return err
	}
	return nil
}

func (sc *SemanticAnalyzer) VisitRule(r *ast.Rule) error {
	sc.pushScope()

	for _, stmt := range r.Statements {
		if err := stmt.(ast.Node).Accept(sc); err != nil {
			return err
		}
	}

	sc.popScope()
	return nil
}

func (sc *SemanticAnalyzer) VisitAssignment(a *ast.Assignment) error {
	if err := a.Expr.(ast.Node).Accept(sc); err != nil {
		return err
	}

	a.Annotate(sc.inferType(a.Expr))
	if err := sc.declareSymbol(a.Name, a.Type, "var", nil, a.Span); err != nil {
		return err
	}

	return nil
}

func (sc *SemanticAnalyzer) VisitEmitStatement(es *ast.EmitStatement) error {
	for _, arg := range es.Args {
		if err := arg.(ast.Node).Accept(sc); err != nil {
			sc.addError(err)
		}
	}

	if es.Condition != nil {
		if err := es.Condition.(ast.Node).Accept(sc); err != nil {
			sc.addError(err)
		}
	}

	signalSym, err := sc.lookupSymbol(es.Signal, es.GetSpan())
	if err != nil {
		sc.addError(errorAt(es.Span, "undefined signal '%s'", es.Signal))
		return nil
	}

	if signalSym.Kind != "signal" {
		sc.addError(errorAt(es.Span, "'%s' is not a signal", es.Signal))
		return nil
	}

	signal := signalSym.Value.(*ast.Signal)
	if len(es.Args) != len(signal.Params) {
		sc.addError(errorAt(es.Span, "signal '%s' expects %d arguments, got %d", es.Signal, len(signal.Params), len(es.Args)))
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
	if err := be.Left.(ast.Node).Accept(sc); err != nil {
		return err
	}
	if err := be.Right.(ast.Node).Accept(sc); err != nil {
		return err
	}

	// check operand types
	leftType := sc.inferType(be.Left)
	rightType := sc.inferType(be.Right)

	switch be.Op {
	case "+":
		valid := sc.requireType(leftType, "number", "string") &&
			sc.requireType(rightType, "number", "string") &&
			sc.typesCompatible(leftType, rightType)

		if !valid {
			return errorAt(be, "operator '%s' cannot be applied to %s and %s", be.Op, leftType, rightType)
		}
		be.Annotate(leftType)
	case "-", "*", "/", "%":
		lok := sc.requireType(leftType, "number")
		rok := sc.requireType(rightType, "number")
		if !(lok && rok) {
			return errorAt(be, "operator '%s' cannot be applied to %s and %s", be.Op, leftType, rightType)
		}
		be.Annotate("number")
	case "==", "!=":
		if !sc.typesCompatible(leftType, rightType) {
			sc.addError(errorAt(be, "cannot compare %s with %s using '%s'", leftType, rightType, be.Op))
		}
		be.Annotate("bool")
	case "<", "<=", ">", ">=":
		lok := sc.requireType(leftType, "number")
		rok := sc.requireType(rightType, "number")
		if !(lok && rok) {
			return errorAt(be, "operator '%s' cannot be applied to %s and %s", be.Op, leftType, rightType)
		}
		be.Annotate("bool")
	case "&&", "||":
		lok := sc.requireType(leftType, "bool")
		rok := sc.requireType(rightType, "bool")
		if !(lok && rok) {
			return errorAt(be, "operator '%s' cannot be applied to %s and %s", be.Op, leftType, rightType)
		}
		be.Annotate("bool")
	case "in":
		if !sc.requireType(rightType, "array") {
			return errorAt(be, "operator '%s' cannot be applied to type %s", be.Op, rightType)
		}
		be.Annotate("bool")
	default:
		sc.addError(errorAt(be, "unknown binary operator '%s'", be.Op))
	}

	return nil
}

func (sc *SemanticAnalyzer) VisitUnaryExpr(ue *ast.UnaryExpr) error {
	if err := ue.Expr.(ast.Node).Accept(sc); err != nil {
		return err
	}

	ty := sc.inferType(ue.Expr)
	switch ue.Op {
	case "!":
		if !sc.requireType(ty, "bool") {
			sc.addError(errorAt(ue, "operand of '!' must be a boolean, got %s", ty))
		}
		ue.Annotate("bool")
	case "-":
		if !sc.requireType(ty, "number") {
			sc.addError(errorAt(ue, "operand of '-' must be a number, got %s", ty))
		}
		ue.Annotate("number")
	default:
		sc.addError(errorAt(ue, "unknown unary operator '%s'", ue.Op))
	}

	return nil
}

func (sc *SemanticAnalyzer) VisitIdentifier(id *ast.Identifier) error {
	if _, err := sc.lookupSymbol(id.Name, id.Span); err != nil {
		return err
	}
	return nil
}

func (sc *SemanticAnalyzer) VisitNumberLiteral(nl *ast.NumberLiteral) error {
	if strings.Contains(nl.Value, ".") {
		val, err := strconv.ParseFloat(nl.Value, 64)
		if err != nil {
			return err
		}
		nl.Kind = ast.FloatNumber
		nl.Float = val
		return nil
	}

	val, err := strconv.ParseInt(nl.Value, 10, 64)
	if err != nil {
		return err
	}

	nl.Kind = ast.IntNumber
	nl.Int = val
	return nil
}

func (sc *SemanticAnalyzer) VisitStringLiteral(sl *ast.StringLiteral) error {
	return nil
}

func (sc *SemanticAnalyzer) VisitBoolLiteral(bl *ast.BoolLiteral) error {
	return nil
}

func (sc *SemanticAnalyzer) VisitConditionalExpr(ce *ast.ConditionalExpr) error {
	if err := ce.Cond.(ast.Node).Accept(sc); err != nil {
		return err
	}
	condType := sc.inferType(ce.Cond)
	if !sc.requireType(condType, "bool") {
		sc.addError(errorAt(ce, "condition in conditional expression must be boolean, got %s", condType))
	}

	if err := ce.Then.(ast.Node).Accept(sc); err != nil {
		return err
	}
	if err := ce.Else.(ast.Node).Accept(sc); err != nil {
		return err
	}

	thenType := sc.inferType(ce.Then)
	elseType := sc.inferType(ce.Else)

	if thenType != elseType {
		sc.addError(errorAt(ce, "then and else branches of conditional expression must have compatible types, got %s and %s", thenType, elseType))
	}

	ce.Annotate(thenType)
	return nil
}

func (sc *SemanticAnalyzer) VisitSelectorExpr(se *ast.SelectorExpr) error {
	switch e := se.Expr.(type) {
	case *ast.Identifier:
		sym, err := sc.lookupSymbol(e.Name, se.GetSpan())
		if err != nil {
			return err
		}
		baseType := sc.inferType(se.Expr)
		if !sc.requireType(sym.Type, "object") {
			return errorAt(se, "cannot access property '%s' on type %s (%s)", se.Field, baseType, e.Name)
		}
	case *ast.IndexExpr:
		if err := se.Expr.(ast.Node).Accept(sc); err != nil {
			return err
		}
	case *ast.SelectorExpr:
		if err := se.Expr.(ast.Node).Accept(sc); err != nil {
			return err
		}
	default:
		return errorAt(se, "cannot access property '%s' on type %s", se.Field, sc.inferType(se.Expr))
	}
	return nil
}

func (sc *SemanticAnalyzer) VisitIndexExpr(ie *ast.IndexExpr) error {
	if err := ie.Expr.(ast.Node).Accept(sc); err != nil {
		return err
	}
	if err := ie.Index.(ast.Node).Accept(sc); err != nil {
		return err
	}

	baseType := sc.inferType(ie.Expr)
	if !sc.requireType(baseType, "object", "array") {
		return errorAt(ie, "cannot index on base type %s", baseType)
	}

	indexType := sc.inferType(ie.Index)
	if !sc.requireType(indexType, "string", "number") {
		return errorAt(ie, "cannot index with type %s", indexType)
	}

	return nil
}

func (sc *SemanticAnalyzer) VisitArrayLiteral(al *ast.ArrayLiteral) error {
	for _, elem := range al.Elements {
		if err := elem.(ast.Node).Accept(sc); err != nil {
			sc.addError(err)
		}
	}
	return nil
}

func (sc *SemanticAnalyzer) VisitParam(p *ast.Param) error {
	return nil
}

func (sc *SemanticAnalyzer) VisitCallExpr(ce *ast.CallExpr) error {
	if !registry.Exists(ce.Namespace, ce.Name) {
		return errorAt(ce, "undefined function '%s'", ce.FullPath())
	}

	fn, _ := registry.Get(ce.Namespace, ce.Name)
	if len(ce.Args) != len(fn.Args) {
		return errorAt(ce, "function '%s' expects %d arguments, got %d",
			fn.FullPath(), len(fn.Args), len(ce.Args))
	}

	for i, arg := range ce.Args {
		if err := arg.(ast.Node).Accept(sc); err != nil {
			return err
		}

		argType := sc.inferType(arg)
		fnArg := fn.Args[i]

		// fn argument does not expect a strict type (any)
		if fnArg.Types == nil {
			continue
		}

		if !sc.requireType(argType, fnArg.Types...) {
			return errorAt(
				arg,
				"%v requires '%v' to be of type: %v",
				fn.CompactSignature(),
				fnArg.Name,
				strings.Join(fnArg.Types, " | "),
			)
		}
	}

	ce.Annotate(fn.ReturnType)
	return nil
}
