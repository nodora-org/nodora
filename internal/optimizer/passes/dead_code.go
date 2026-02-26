package passes

import (
	"slices"

	"nodora.org/nodora/pkg/nir"
)

type DeadCodeElimination struct{}

func NewDeadCodeElimination() *DeadCodeElimination {
	return &DeadCodeElimination{}
}

func (dce *DeadCodeElimination) Run(p *nir.Program) (bool, error) {
	changed := false

	// collect used signals by scanning all rules
	usedSignals := dce.collectUsedSignals(p)

	// remove unused signals
	for key := range p.Signals {
		if !usedSignals[key] {
			delete(p.Signals, key)
			changed = true
		}
	}

	// eliminate dead operations
	for key := range p.Rules {
		r := p.Rules[key]
		if dce.eliminateDeadOperations(&r) {
			changed = true
		}
		p.Rules[key] = r
	}

	return changed, nil
}

func (dce *DeadCodeElimination) collectUsedSignals(p *nir.Program) map[string]bool {
	usedSignals := make(map[string]bool)
	for _, r := range p.Rules {
		for _, op := range r.Ops {
			if op.Kind == nir.OpEmit {
				if expr, ok := op.Args[0].Expr.(*nir.SignalExpr); ok {
					usedSignals[expr.Name] = true
				}
			}
		}
	}
	return usedSignals
}

func (dce *DeadCodeElimination) eliminateDeadOperations(r *nir.Rule) bool {
	if len(r.Ops) == 0 {
		return false
	}

	// start with slots used by outputs
	usedSlots := make(map[int]bool)
	for _, output := range r.Outputs {
		usedSlots[output.Sym] = true
	}

	// work backwards to find live operations and track slot dependencies
	liveOps := make([]nir.Op, 0, len(r.Ops))

	for i := len(r.Ops) - 1; i >= 0; i-- {
		op := &r.Ops[i]
		if dce.isLiveOp(op, usedSlots) {
			for _, arg := range op.Args {
				dce.collectSymReads(&arg, usedSlots)
			}
			liveOps = append(liveOps, *op)
		}
	}

	// reverse to restore original order
	slices.Reverse(liveOps)

	changed := len(liveOps) != len(r.Ops)
	r.Ops = liveOps
	return changed
}

func (dce *DeadCodeElimination) collectSymReads(e *nir.RawExpr, usedSlots map[int]bool) {
	switch e := e.Expr.(type) {
	case *nir.SymExpr:
		usedSlots[e.Index] = true
	case *nir.IdxExpr:
		dce.collectSymReads(&e.From, usedSlots)
		dce.collectSymReads(&e.Index, usedSlots)
	case *nir.SelExpr:
		dce.collectSymReads(&e.From, usedSlots)
		for _, expr := range e.Exprs {
			dce.collectSymReads(&expr, usedSlots)
		}
	case *nir.ArrExpr:
		for _, arg := range e.Value {
			dce.collectSymReads(&arg, usedSlots)
		}
	case *nir.ObjExpr:
		for _, prop := range e.Value {
			dce.collectSymReads(&prop, usedSlots)
		}
	case *nir.CallExpr:
		for _, arg := range e.Args {
			dce.collectSymReads(&arg, usedSlots)
		}
	case *nir.SignalExpr:
		for _, arg := range e.Args {
			dce.collectSymReads(&arg, usedSlots)
		}
		if e.When != nil {
			dce.collectSymReads(e.When, usedSlots)
		}
	case *nir.LambdaExpr:
		for i := range e.Ops {
			op := &e.Ops[i]
			for _, arg := range op.Args {
				dce.collectSymReads(&arg, usedSlots)
			}
		}
	}
}

func (dce *DeadCodeElimination) isLiveOp(op *nir.Op, usedSlots map[int]bool) bool {
	// ops with side effects are always live
	if op.Out == nil {
		return true
	}
	return usedSlots[*op.Out] // live if its output slot is used
}
