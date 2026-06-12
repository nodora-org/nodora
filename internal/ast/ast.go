package ast

import (
	"fmt"
	"strings"

	"nodora.org/nodora/pkg/types"
)

type Node interface {
	Accept(visitor Visitor) error
}

type Typed struct {
	Type types.Type
}

func (t *Typed) Annotate(ty types.Type) {
	t.Type = ty
}

type Program struct {
	Decls []Node
}

func (p *Program) Accept(visitor Visitor) error {
	return visitor.VisitProgram(p)
}

func (p *Program) String() string {
	decls := make([]string, len(p.Decls))
	for i, decl := range p.Decls {
		decls[i] = fmt.Sprintf("%v", decl)
	}
	return strings.Join(decls, "\n\n")
}

type Signal struct {
	Span
	Name   string
	Params []Param
}

func (s *Signal) Accept(visitor Visitor) error {
	return visitor.VisitSignal(s)
}

func (s *Signal) String() string {
	params := make([]string, len(s.Params))
	for i, p := range s.Params {
		params[i] = p.Name
	}
	return fmt.Sprintf("signal %s(%s)", s.Name, strings.Join(params, ", "))
}

type Const struct {
	Typed
	Span
	Name  string
	Value Expr
}

func (c *Const) Accept(visitor Visitor) error {
	return visitor.VisitConst(c)
}

func (c *Const) String() string {
	return fmt.Sprintf("const %s = %v", c.Name, c.Value)
}

type Rule struct {
	Span
	Name       string
	Statements []Statement
}

func (r *Rule) Accept(visitor Visitor) error {
	return visitor.VisitRule(r)
}

func (r *Rule) String() string {
	stmts := make([]string, len(r.Statements))
	for i, stmt := range r.Statements {
		stmts[i] = fmt.Sprintf("\t%v", stmt)
	}
	return fmt.Sprintf("rule %s {\n%s\n}", r.Name, strings.Join(stmts, "\n"))
}

type Param struct {
	Span
	Name string
}

func (p Param) String() string {
	return p.Name
}

func (p *Param) Accept(visitor Visitor) error {
	return visitor.VisitParam(p)
}

type Statement interface {
	Node
	WithSpan
}

type Assignment struct {
	Typed
	Span
	Name  string
	Expr  Expr
	IsOut bool
}

func (a *Assignment) String() string {
	if a.IsOut {
		return fmt.Sprintf("out %s = %v", a.Name, a.Expr)
	}
	return fmt.Sprintf("%s = %v", a.Name, a.Expr)
}

func (a *Assignment) Accept(visitor Visitor) error {
	return visitor.VisitAssignment(a)
}

type EmitStatement struct {
	Span
	Signal    string
	Args      []Expr
	Condition Expr
}

func (es *EmitStatement) Accept(visitor Visitor) error {
	return visitor.VisitEmitStatement(es)
}

func (es *EmitStatement) String() string {
	args := make([]string, len(es.Args))
	for i, arg := range es.Args {
		args[i] = fmt.Sprintf("%v", arg)
	}
	if es.Condition != nil {
		return fmt.Sprintf("emit %s(%s) when %v", es.Signal, strings.Join(args, ", "), es.Condition)
	}
	return fmt.Sprintf("emit %s(%s)", es.Signal, strings.Join(args, ", "))
}

type Expr interface {
	Node
	WithSpan
}

type BinaryExpr struct {
	Typed
	Span
	Left  Expr
	Op    string
	Right Expr
}

func (be *BinaryExpr) Accept(visitor Visitor) error {
	return visitor.VisitBinaryExpr(be)
}

func (be *BinaryExpr) String() string {
	return fmt.Sprintf("(%v %s %v)", be.Left, be.Op, be.Right)
}

type UnaryExpr struct {
	Typed
	Span
	Op   string
	Expr Expr
}

func (ue *UnaryExpr) Accept(visitor Visitor) error {
	return visitor.VisitUnaryExpr(ue)
}

func (ue *UnaryExpr) String() string {
	return fmt.Sprintf("%s%v", ue.Op, ue.Expr)
}

type SelectorExpr struct {
	Typed
	Span
	Expr  Expr
	Field string
}

func (se *SelectorExpr) Accept(visitor Visitor) error {
	return visitor.VisitSelectorExpr(se)
}

func (se *SelectorExpr) String() string {
	return fmt.Sprintf("%v.%s", se.Expr, se.Field)
}

type IndexExpr struct {
	Typed
	Span
	Expr  Expr
	Index Expr
}

func (ie *IndexExpr) Accept(visitor Visitor) error {
	return visitor.VisitIndexExpr(ie)
}

func (ie *IndexExpr) String() string {
	return fmt.Sprintf("%v[%v]", ie.Expr, ie.Index)
}

type Identifier struct {
	Typed
	Span
	Name string
}

func (id *Identifier) Accept(visitor Visitor) error {
	return visitor.VisitIdentifier(id)
}

func (id *Identifier) String() string {
	return id.Name
}

type NumberLiteral struct {
	Typed
	Span
	Raw   string
	Value float64
}

func (nl *NumberLiteral) Accept(visitor Visitor) error {
	return visitor.VisitNumberLiteral(nl)
}

func (nl *NumberLiteral) String() string {
	return nl.Raw
}

func (nl *NumberLiteral) IsInt() bool {
	return !strings.Contains(nl.Raw, ".")
}

type StringLiteral struct {
	Typed
	Span
	Value string
}

func (sl *StringLiteral) Accept(visitor Visitor) error {
	return visitor.VisitStringLiteral(sl)
}

func (sl *StringLiteral) String() string {
	return sl.Value
}

type BoolLiteral struct {
	Typed
	Span
	Value bool
}

func (bl *BoolLiteral) Accept(visitor Visitor) error {
	return visitor.VisitBoolLiteral(bl)
}

func (bl *BoolLiteral) String() string {
	return fmt.Sprintf("%v", bl.Value)
}

type ConditionalExpr struct {
	Typed
	Span
	Cond Expr
	Then Expr
	Else Expr
}

func (ce *ConditionalExpr) Accept(visitor Visitor) error {
	return visitor.VisitConditionalExpr(ce)
}

func (ce *ConditionalExpr) String() string {
	return fmt.Sprintf("%v ? %v : %v", ce.Cond, ce.Then, ce.Else)
}

type ArrayLiteral struct {
	Typed
	Span
	Elements []Expr
}

func (al *ArrayLiteral) Accept(visitor Visitor) error {
	return visitor.VisitArrayLiteral(al)
}

func (al *ArrayLiteral) String() string {
	elems := make([]string, len(al.Elements))
	for i, elem := range al.Elements {
		elems[i] = fmt.Sprintf("%v", elem)
	}
	return fmt.Sprintf("[%s]", strings.Join(elems, ", "))
}

type ObjectLiteral struct {
	Typed
	Span
	Properties []ObjectProperty
}

func (ol *ObjectLiteral) String() string {
	props := make([]string, len(ol.Properties))
	for i, prop := range ol.Properties {
		props[i] = prop.String()
	}
	return fmt.Sprintf("{%s}", strings.Join(props, ", "))
}

type ObjectProperty struct {
	Typed
	Span
	Key   string
	Value Expr
}

func (op *ObjectProperty) String() string {
	return fmt.Sprintf("%s: %s", op.Key, op.Value)
}

func (ol *ObjectLiteral) Accept(visitor Visitor) error {
	return visitor.VisitObjectLiteral(ol)
}

type CallExpr struct {
	Typed
	Span
	Namespace string
	Name      string
	Args      []Expr
}

func (ce *CallExpr) FullPath() string {
	if ce.Namespace == "" {
		return ce.Name
	}
	return ce.Namespace + "::" + ce.Name
}

func (c *CallExpr) Accept(visitor Visitor) error {
	return visitor.VisitCallExpr(c)
}

func (ce *CallExpr) String() string {
	args := make([]string, len(ce.Args))
	for i, arg := range ce.Args {
		args[i] = fmt.Sprintf("%v", arg)
	}
	return fmt.Sprintf("%s(%s)", ce.FullPath(), strings.Join(args, ", "))
}

type LambdaExpr struct {
	Typed
	Span
	Params []Param
	Body   Expr
}

func (le *LambdaExpr) Accept(visitor Visitor) error {
	return visitor.VisitLambdaExpr(le)
}

func (le *LambdaExpr) String() string {
	params := make([]string, len(le.Params))
	for i, p := range le.Params {
		params[i] = p.Name
	}
	return fmt.Sprintf("|%s| %v", strings.Join(params, ", "), le.Body)
}

type ReservedObject struct {
	Typed
	Span
	Name string
}

func (ro *ReservedObject) Accept(visitor Visitor) error {
	return nil
}

type MatchArm struct {
	Span
	Pattern Expr
	Guard   Expr
	Body    Expr
}

func (a *MatchArm) IsCatchAll() bool {
	_, isIdent := a.Pattern.(*Identifier)
	return isIdent && a.Guard == nil
}

func (a *MatchArm) BindingName() string {
	id, ok := a.Pattern.(*Identifier)
	if !ok || id.Name == "_" {
		return ""
	}
	return id.Name
}

type MatchExpr struct {
	Typed
	Span
	Value Expr
	Arms  []*MatchArm
}

func (m *MatchExpr) Accept(visitor Visitor) error {
	return visitor.VisitMatchExpr(m)
}

func (m *MatchExpr) String() string {
	arms := make([]string, len(m.Arms))
	for i, arm := range m.Arms {
		if arm.Guard != nil {
			arms[i] = fmt.Sprintf("%v when %v => %v", arm.Pattern, arm.Guard, arm.Body)
		} else {
			arms[i] = fmt.Sprintf("%v => %v", arm.Pattern, arm.Body)
		}
	}
	return fmt.Sprintf("match %v { %s }", m.Value, strings.Join(arms, ", "))
}
