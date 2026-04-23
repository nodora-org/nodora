package ast

import "fmt"

type Position struct {
	Line int `json:"line"`
	Col  int `json:"col"`
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Col)
}

type WithSpan interface {
	GetSpan() Span
}

type Span struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

func (s *Span) SetSpan(startLine, startCol, endLine, endCol int) {
	s.Start = Position{Line: startLine, Col: startCol}
	s.End = Position{Line: endLine, Col: endCol}
}

func (s Span) Merge(other Span) Span {
	return Span{
		Start: s.Start,
		End:   other.End,
	}
}

func (s Span) String() string {
	return s.Start.String()
}

func (s Span) GetSpan() Span {
	return s
}
