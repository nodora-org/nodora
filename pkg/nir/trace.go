package nir

import (
	"nodora.org/nodora/pkg/core"
)

type NodeKind string

const (
	NodeRoot   NodeKind = "root"
	NodeOp     NodeKind = "op"
	NodeExpr   NodeKind = "expr"
	NodeLambda NodeKind = "lambda"
)

type TraceNode struct {
	Kind NodeKind
	Op   *Op
	Expr Expr
	Args []core.Value
}

func (n TraceNode) Text() string {
	switch n.Kind {
	case NodeOp:
		if n.Op != nil {
			return n.Op.String()
		}
	case NodeExpr, NodeLambda:
		if n.Expr != nil {
			return RawExpr{Expr: n.Expr}.String()
		}
	}
	return ""
}

func (n TraceNode) Lambda() (*LambdaExpr, bool) {
	if n.Kind != NodeLambda {
		return nil, false
	}
	le, ok := n.Expr.(*LambdaExpr)
	return le, ok
}

func (n TraceNode) Slot() (int, bool) {
	if n.Kind != NodeOp || n.Op == nil || n.Op.Out == nil {
		return 0, false
	}
	return *n.Op.Out, true
}

type Tracer interface {
	Enter(node TraceNode)
	Exit(value core.Value, err error)
}

type Event struct {
	TraceNode
	Value    core.Value
	Err      error
	Children []*Event
}

type Recorder struct {
	root  Event
	stack []*Event
}

func NewRecorder() *Recorder {
	r := &Recorder{root: Event{TraceNode: TraceNode{Kind: NodeRoot}}}
	r.stack = []*Event{&r.root}
	return r
}

// synthetic top-level event
func (r *Recorder) Root() *Event {
	return &r.root
}

func (r *Recorder) Enter(node TraceNode) {
	ev := &Event{TraceNode: node}
	parent := r.stack[len(r.stack)-1]
	parent.Children = append(parent.Children, ev)
	r.stack = append(r.stack, ev)
}

func (r *Recorder) Exit(value core.Value, err error) {
	if len(r.stack) <= 1 {
		return // never happens via the instrumented paths
	}
	ev := r.stack[len(r.stack)-1]
	ev.Value = value
	ev.Err = err
	r.stack = r.stack[:len(r.stack)-1]
}
