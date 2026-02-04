package semantics

import (
	"fmt"

	"nodora.org/nodora/internal/ast"
)

type Scope struct {
	symbols map[string]*Symbol
	parent  *Scope
}

type Symbol struct {
	Name  string
	Type  string
	Kind  string
	Value any
}

type SemanticAnalyzer struct {
	scopes []*Scope
	errors []error
}

func NewSemanticAnalyzer() *SemanticAnalyzer {
	return &SemanticAnalyzer{
		scopes: []*Scope{newScope(nil)}, // global scope
		errors: []error{},
	}
}

func newScope(parent *Scope) *Scope {
	symbols := make(map[string]*Symbol)
	symbols["input"] = &Symbol{
		Name:  "input",
		Type:  "object",
		Kind:  "namespace",
		Value: make(map[string]any),
	}
	return &Scope{
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

func (sc *SemanticAnalyzer) currentScope() *Scope {
	return sc.scopes[len(sc.scopes)-1]
}

func (sc *SemanticAnalyzer) isReservedNamespace(name string) bool {
	return name == "input"
}

func (sc *SemanticAnalyzer) declareSymbol(name, typ, kind string, value any) error {
	if sc.isReservedNamespace(name) {
		return fmt.Errorf("symbol '%s' is a reserved namespace", name)
	}
	scope := sc.currentScope()
	if _, exists := scope.symbols[name]; exists {
		return fmt.Errorf("symbol '%s' already declared in this scope", name)
	}
	scope.symbols[name] = &Symbol{Name: name, Type: typ, Kind: kind, Value: value}
	return nil
}

func (sc *SemanticAnalyzer) lookupSymbol(name string) (*Symbol, error) {
	for i := len(sc.scopes) - 1; i >= 0; i-- {
		if sym, exists := sc.scopes[i].symbols[name]; exists {
			return sym, nil
		}
	}
	return nil, fmt.Errorf("undefined symbol '%s'", name)
}

func (sc *SemanticAnalyzer) addError(err error) {
	sc.errors = append(sc.errors, err)
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
		if sym, err := sc.lookupSymbol(e.Name); err == nil {
			return sym.Type
		}
		return "unknown" // symbol not found
	case *ast.BinaryExpr:
		if e.Type != "" {
			return e.Type
		}
		switch e.Op {
		case "+", "-", "*", "/", "%":
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
	default:
		return "unknown"
	}
}

func (sc *SemanticAnalyzer) requireType(actualType string, expectedType string) bool {
	return actualType == expectedType || actualType == "unknown"
}

func (sc *SemanticAnalyzer) Analyze(program *ast.Program) []error {
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

	return sc.errors
}

func (sc *SemanticAnalyzer) VisitProgram(p *ast.Program) error {
	return nil // handled in Analyze
}

func (sc *SemanticAnalyzer) VisitSignal(s *ast.Signal) error {
	if err := sc.declareSymbol(s.Name, "", "signal", s); err != nil {
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
	if err := sc.declareSymbol(a.Name, a.Type, "var", nil); err != nil {
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

	signalSym, err := sc.lookupSymbol(es.Signal)
	if err != nil {
		sc.addError(fmt.Errorf("undefined signal '%s'", es.Signal))
		return nil
	}

	if signalSym.Kind != "signal" {
		sc.addError(fmt.Errorf("'%s' is not a signal", es.Signal))
		return nil
	}

	signal := signalSym.Value.(*ast.Signal)
	if len(es.Args) != len(signal.Params) {
		sc.addError(fmt.Errorf("signal '%s' expects %d arguments, got %d", es.Signal, len(signal.Params), len(es.Args)))
		return nil
	}

	for _, arg := range es.Args {
		if ident, ok := arg.(*ast.Identifier); ok {
			if _, err := sc.lookupSymbol(ident.Name); err != nil {
				sc.addError(fmt.Errorf("undefined symbol '%s' in signal emit argument", ident.Name))
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
	case "+", "-", "*", "/", "%":
		// arithmetic operators require both operands to be numbers
		lok := sc.requireType(leftType, "number")
		rok := sc.requireType(rightType, "number")
		if !(lok && rok) {
			return fmt.Errorf("operator '%s' cannot be applied to %s and %s", be.Op, leftType, rightType)
		}
		be.Annotate("number")
	case "==", "!=":
		if leftType != rightType && leftType != "unknown" && rightType != "unknown" {
			sc.addError(fmt.Errorf("cannot compare %s with %s using '%s'", leftType, rightType, be.Op))
		}
		be.Annotate("bool")
	case "<", "<=", ">", ">=":
		// comparison operators require compatible numeric types
		lok := sc.requireType(leftType, "number")
		rok := sc.requireType(rightType, "number")
		if !(lok && rok) {
			return fmt.Errorf("operator '%s' cannot be applied to %s and %s", be.Op, leftType, rightType)
		}
		be.Annotate("bool")
	case "&&", "||":
		// logical operators require both operands to be booleans
		lok := sc.requireType(leftType, "bool")
		rok := sc.requireType(rightType, "bool")
		if !(lok && rok) {
			return fmt.Errorf("operator '%s' cannot be applied to %s and %s", be.Op, leftType, rightType)
		}
		be.Annotate("bool")
	case "in":
		// right operand should be an array
		if !sc.requireType(rightType, "array") {
			return fmt.Errorf("operator '%s' cannot be applied to type %s", be.Op, rightType)
		}
		be.Annotate("bool")
	default:
		sc.addError(fmt.Errorf("unknown binary operator '%s'", be.Op))
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
		// logical NOT requires boolean operand
		if ty != "bool" && ty != "unknown" {
			sc.addError(fmt.Errorf("operand of '!' must be a boolean, got %s", ty))
		}
		ue.Annotate("bool")
	case "-":
		if ty != "number" && ty != "unknown" {
			sc.addError(fmt.Errorf("operand of '-' must be a number, got %s", ty))
		}
		ue.Annotate("number")
	default:
		sc.addError(fmt.Errorf("unknown unary operator '%s'", ue.Op))
	}

	return nil
}

func (sc *SemanticAnalyzer) VisitIdentifier(id *ast.Identifier) error {
	if _, err := sc.lookupSymbol(id.Name); err != nil {
		return err
	}
	return nil
}

func (sc *SemanticAnalyzer) VisitNumberLiteral(nl *ast.NumberLiteral) error {
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
	if condType != "bool" {
		sc.addError(fmt.Errorf("condition in conditional expression must be boolean, got %s", condType))
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
		sc.addError(fmt.Errorf("then and else branches of conditional expression must have compatible types, got %s and %s", thenType, elseType))
	}

	ce.Annotate(thenType)
	return nil
}

func (sc *SemanticAnalyzer) VisitSelectorExpr(se *ast.SelectorExpr) error {
	switch e := se.Expr.(type) {
	case *ast.Identifier:
		sym, err := sc.lookupSymbol(e.Name)
		if err != nil {
			return err
		}
		baseType := sc.inferType(se.Expr)
		if sym.Type != "unknown" && sym.Type != "object" {
			return fmt.Errorf("cannot access property '%s' on type %s (%s)", se.Field, baseType, e.Name)
		}
	case *ast.SelectorExpr:
		if err := se.Expr.(ast.Node).Accept(sc); err != nil {
			return err
		}
	default:
		return fmt.Errorf("cannot access property '%s' on type %s", se.Field, sc.inferType(se.Expr))
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
