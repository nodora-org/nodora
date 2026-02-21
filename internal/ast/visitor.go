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
	VisitIndexExpr(ie *IndexExpr) error
	VisitIdentifier(id *Identifier) error
	VisitNumberLiteral(nl *NumberLiteral) error
	VisitStringLiteral(sl *StringLiteral) error
	VisitBoolLiteral(bl *BoolLiteral) error
	VisitConditionalExpr(ce *ConditionalExpr) error
	VisitArrayLiteral(al *ArrayLiteral) error
	VisitObjectLiteral(ol *ObjectLiteral) error
	VisitCallExpr(ce *CallExpr) error
	VisitParam(p *Param) error
}

type BaseVisitor struct{}

func (bv *BaseVisitor) VisitProgram(p *Program) error {
	for _, decl := range p.Decls {
		if err := decl.Accept(bv); err != nil {
			return err
		}
	}
	return nil
}

func (bv *BaseVisitor) VisitSignal(s *Signal) error { return nil }

func (bv *BaseVisitor) VisitRule(r *Rule) error {
	for _, stmt := range r.Statements {
		if err := stmt.(Node).Accept(bv); err != nil {
			return err
		}
	}
	return nil
}

func (bv *BaseVisitor) VisitAssignment(a *Assignment) error {
	return a.Expr.(Node).Accept(bv)
}

func (bv *BaseVisitor) VisitEmitStatement(es *EmitStatement) error {
	for _, arg := range es.Args {
		if err := arg.(Node).Accept(bv); err != nil {
			return err
		}
	}
	if es.Condition != nil {
		if err := (es.Condition).(Node).Accept(bv); err != nil {
			return err
		}
	}
	return nil
}

func (bv *BaseVisitor) VisitBinaryExpr(be *BinaryExpr) error {
	if err := be.Left.(Node).Accept(bv); err != nil {
		return err
	}
	if err := be.Right.(Node).Accept(bv); err != nil {
		return err
	}
	return nil
}

func (bv *BaseVisitor) VisitUnaryExpr(ue *UnaryExpr) error {
	return ue.Expr.(Node).Accept(bv)
}

func (bv *BaseVisitor) VisitSelectorExpr(se *SelectorExpr) error {
	if err := se.Expr.(Node).Accept(bv); err != nil {
		return err
	}
	return nil
}

func (bv *BaseVisitor) VisitIndexExpr(ie *IndexExpr) error {
	if err := ie.Expr.(Node).Accept(bv); err != nil {
		return err
	}
	return ie.Index.(Node).Accept(bv)
}
func (bv *BaseVisitor) VisitIdentifier(id *Identifier) error { return nil }

func (bv *BaseVisitor) VisitNumberLiteral(nl *NumberLiteral) error { return nil }

func (bv *BaseVisitor) VisitStringLiteral(sl *StringLiteral) error { return nil }

func (bv *BaseVisitor) VisitBoolLiteral(bl *BoolLiteral) error { return nil }

func (bv *BaseVisitor) VisitConditionalExpr(ce *ConditionalExpr) error {
	if err := ce.Cond.(Node).Accept(bv); err != nil {
		return err
	}
	if err := ce.Then.(Node).Accept(bv); err != nil {
		return err
	}
	if err := ce.Else.(Node).Accept(bv); err != nil {
		return err
	}
	return nil
}

func (bv *BaseVisitor) VisitArrayLiteral(al *ArrayLiteral) error {
	for _, elem := range al.Elements {
		if err := elem.(Node).Accept(bv); err != nil {
			return err
		}
	}
	return nil
}

func (bv *BaseVisitor) VisitObjectLiteral(ol *ObjectLiteral) error {
	for _, prop := range ol.Properties {
		if err := prop.Value.(Node).Accept(bv); err != nil {
			return err
		}
	}
	return nil
}

func (bv *BaseVisitor) VisitCallExpr(ce *CallExpr) error {
	for _, arg := range ce.Args {
		if err := arg.(Node).Accept(bv); err != nil {
			return err
		}
	}
	return nil
}

func (bv *BaseVisitor) VisitParam(p *Param) error { return nil }
