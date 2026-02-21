package passes

import (
	"nodora.org/nodora/pkg/nir"
)

type SymbolRemap struct{}

func NewSymbolRemap() *SymbolRemap {
	return &SymbolRemap{}
}

func (sr *SymbolRemap) Run(p *nir.Program) error {
	for key := range p.Rules {
		r := p.Rules[key]
		sr.remapSlots(&r)
		p.Rules[key] = r
	}
	return nil
}

func (sr *SymbolRemap) remapSlots(r *nir.Rule) {
	slotMap := make(map[int]int)
	nextIdx := 1

	// remap operation outputs
	for i := range r.Ops {
		op := &r.Ops[i]
		if op.Out != nil {
			slotMap[*op.Out] = nextIdx
			*op.Out = nextIdx
			nextIdx++
		}
		// remap sym slots in arguments
		for i := range op.Args {
			sr.remapSymExpr(&op.Args[i], slotMap)
		}
	}

	// remap outputs
	for name, output := range r.Outputs {
		if newIdx, ok := slotMap[output.Sym]; ok {
			output.Sym = newIdx
			r.Outputs[name] = output
		}
	}

	// update symslots (highest used slot + 1)
	r.Symslots = nextIdx
}

func (sr *SymbolRemap) remapSymExpr(arg *nir.RawExpr, slotMap map[int]int) {
	if arg == nil || arg.Expr == nil {
		return
	}

	switch e := arg.Expr.(type) {
	case *nir.SymExpr:
		if newIdx, ok := slotMap[e.Index]; ok {
			e.Index = newIdx
		}
	case *nir.IdxExpr:
		sr.remapSymExpr(&e.From, slotMap)
		sr.remapSymExpr(&e.Index, slotMap)
	case *nir.SelExpr:
		sr.remapSymExpr(&e.From, slotMap)
		for i := range e.Exprs {
			sr.remapSymExpr(&e.Exprs[i], slotMap)
		}
	case *nir.ArrExpr:
		for i := range e.Value {
			sr.remapSymExpr(&e.Value[i], slotMap)
		}
	case *nir.ObjExpr:
		for _, prop := range e.Value {
			sr.remapSymExpr(&prop, slotMap)
		}
	case *nir.SignalExpr:
		for i := range e.Args {
			sr.remapSymExpr(&e.Args[i], slotMap)
		}
		if e.When != nil {
			sr.remapSymExpr(e.When, slotMap)
		}
	}
}
