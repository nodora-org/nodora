package ast

type Visitor interface {
	VisitProgram(p *Program) error
	VisitSignal(s *Signal) error
	VisitRule(r *Rule) error
	VisitAssignment(a *Assignment) error
	VisitEmitStatement(es *EmitStatement) error
	VisitBinaryExpr(be *BinaryExpr) error
	VisitUnaryExpr(ue *UnaryExpr) error
	VisitSelectorExpr(se *SelectorExpr) error
	VisitIdentifier(id *Identifier) error
	VisitNumberLiteral(nl *NumberLiteral) error
	VisitStringLiteral(sl *StringLiteral) error
	VisitBoolLiteral(bl *BoolLiteral) error
	VisitConditionalExpr(ce *ConditionalExpr) error
	VisitArrayLiteral(al *ArrayLiteral) error
	VisitParam(p *Param) error
}

// Program traversal
func VisitProgram(dv Visitor, p *Program) error {
	for _, decl := range p.Decls {
		if err := decl.Accept(dv); err != nil {
			return err
		}
	}
	return nil
}

// Declarations - no children to traverse
func VisitSignal(dv Visitor, s *Signal) error {
	return nil
}

// Rule traversal
func VisitRule(dv Visitor, r *Rule) error {
	for _, stmt := range r.Statements {
		if err := stmt.(Node).Accept(dv); err != nil {
			return err
		}
	}
	return nil
}

// Assignment traversal
func VisitAssignment(dv Visitor, a *Assignment) error {
	return a.Expr.(Node).Accept(dv)
}

// EmitStatement traversal
func VisitEmitStatement(dv Visitor, es *EmitStatement) error {
	for _, arg := range es.Args {
		if err := arg.(Node).Accept(dv); err != nil {
			return err
		}
	}
	if es.Condition != nil {
		if err := (es.Condition).(Node).Accept(dv); err != nil {
			return err
		}
	}
	return nil
}

// BinaryExpr traversal
func VisitBinaryExpr(dv Visitor, be *BinaryExpr) error {
	if err := be.Left.(Node).Accept(dv); err != nil {
		return err
	}
	if err := be.Right.(Node).Accept(dv); err != nil {
		return err
	}
	return nil
}

// UnaryExpr traversal
func VisitUnaryExpr(dv Visitor, ue *UnaryExpr) error {
	return ue.Expr.(Node).Accept(dv)
}

// SelectorExpr traversal
func VisitSelectorExpr(dv Visitor, se *SelectorExpr) error {
	if err := se.Expr.(Node).Accept(dv); err != nil {
		return err
	}
	return nil
}

// Identifier - leaf node
func VisitIdentifier(dv Visitor, id *Identifier) error {
	return nil
}

// Literals - leaf nodes
func VisitNumberLiteral(dv Visitor, nl *NumberLiteral) error {
	return nil
}

func VisitStringLiteral(dv Visitor, sl *StringLiteral) error {
	return nil
}

func VisitBoolLiteral(dv Visitor, bl *BoolLiteral) error {
	return nil
}

// ConditionalExpr traversal
func VisitConditionalExpr(dv Visitor, ce *ConditionalExpr) error {
	if err := ce.Cond.(Node).Accept(dv); err != nil {
		return err
	}
	if err := ce.Then.(Node).Accept(dv); err != nil {
		return err
	}
	if err := ce.Else.(Node).Accept(dv); err != nil {
		return err
	}
	return nil
}

// ArrayLiteral traversal
func VisitArrayLiteral(dv Visitor, al *ArrayLiteral) error {
	for _, elem := range al.Elements {
		if err := elem.(Node).Accept(dv); err != nil {
			return err
		}
	}
	return nil
}

// Param - leaf node
func VisitParam(dv Visitor, p *Param) error {
	return nil
}
