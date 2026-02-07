package ast

import "strings"

type Node interface {
	Accept(visitor Visitor) error
}

type Typed struct {
	Type string
}

func (t *Typed) Annotate(ty string) {
	t.Type = ty
}

type Program struct {
	Decls []Node
}

func (p *Program) Accept(visitor Visitor) error {
	return visitor.VisitProgram(p)
}

type Signal struct {
	Name   string
	Params []Param
}

func (s *Signal) Accept(visitor Visitor) error {
	return visitor.VisitSignal(s)
}

type Rule struct {
	Name       string
	Statements []Statement
}

func (r *Rule) Accept(visitor Visitor) error {
	return visitor.VisitRule(r)
}

type Param struct {
	Name string
}

func (p *Param) Accept(visitor Visitor) error {
	return visitor.VisitParam(p)
}

type Statement interface {
	Node
}

type Assignment struct {
	Typed
	Name  string
	Expr  Expr
	IsOut bool
}

func (a *Assignment) Accept(visitor Visitor) error {
	return visitor.VisitAssignment(a)
}

type EmitStatement struct {
	Signal    string
	Args      []Expr
	Condition Expr
}

func (es *EmitStatement) Accept(visitor Visitor) error {
	return visitor.VisitEmitStatement(es)
}

type Expr interface{}

type BinaryExpr struct {
	Typed
	Left  Expr
	Op    string
	Right Expr
}

func (be *BinaryExpr) Accept(visitor Visitor) error {
	return visitor.VisitBinaryExpr(be)
}

type UnaryExpr struct {
	Typed
	Op   string
	Expr Expr
}

func (ue *UnaryExpr) Accept(visitor Visitor) error {
	return visitor.VisitUnaryExpr(ue)
}

type SelectorExpr struct {
	Expr  Expr
	Field string
}

func (se *SelectorExpr) Accept(visitor Visitor) error {
	return visitor.VisitSelectorExpr(se)
}

func (se *SelectorExpr) Path() string {
	parts := []string{se.Field}
	expr := se.Expr

	for {
		switch e := expr.(type) {
		case *SelectorExpr:
			parts = append([]string{e.Field}, parts...)
			expr = e.Expr
		case *Identifier:
			parts = append([]string{e.Name}, parts...)
			return strings.Join(parts, ".")
		default:
			return ""
		}
	}
}

type IndexExpr struct {
	Expr  Expr
	Index Expr
}

func (ie *IndexExpr) Accept(visitor Visitor) error {
	return visitor.VisitIndexExpr(ie)
}

type Identifier struct {
	Name string
}

func (id *Identifier) Accept(visitor Visitor) error {
	return visitor.VisitIdentifier(id)
}

type NumberKind int

const (
	IntNumber NumberKind = iota
	FloatNumber
)

type NumberLiteral struct {
	Kind  NumberKind
	Value string
	Int   int64
	Float float64
}

func (nl *NumberLiteral) Accept(visitor Visitor) error {
	return visitor.VisitNumberLiteral(nl)
}

func (nl *NumberLiteral) GetValue() (bool, any) {
	switch nl.Kind {
	case FloatNumber:
		return true, nl.Float
	case IntNumber:
		return true, nl.Int
	}
	return false, nil
}

type StringLiteral struct {
	Value string
}

func (sl *StringLiteral) Accept(visitor Visitor) error {
	return visitor.VisitStringLiteral(sl)
}

type BoolLiteral struct {
	Value bool
}

func (bl *BoolLiteral) Accept(visitor Visitor) error {
	return visitor.VisitBoolLiteral(bl)
}

type ConditionalExpr struct {
	Typed
	Cond Expr
	Then Expr
	Else Expr
}

func (ce *ConditionalExpr) Accept(visitor Visitor) error {
	return visitor.VisitConditionalExpr(ce)
}

type ArrayLiteral struct {
	Elements []Expr
}

func (al *ArrayLiteral) Accept(visitor Visitor) error {
	return visitor.VisitArrayLiteral(al)
}
