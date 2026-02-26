package passes

import (
	"nodora.org/nodora/pkg/nir"
)

type SymbolRemap struct {
	slotMap map[int]int
	nextIdx int
}

func NewSymbolRemap() *SymbolRemap {
	return &SymbolRemap{}
}

func (sr *SymbolRemap) Run(p *nir.Program) (bool, error) {
	for key := range p.Rules {
		r := p.Rules[key]
		sr.remapSlots(&r)
		p.Rules[key] = r
	}
	return false, nil
}

func (sr *SymbolRemap) remapSlots(r *nir.Rule) {
	sr.slotMap = make(map[int]int)
	sr.nextIdx = 1

	// remap operation outputs
	for i := range r.Ops {
		sr.remapOp(&r.Ops[i])
	}

	// remap outputs
	for name, output := range r.Outputs {
		if newIdx, ok := sr.slotMap[output.Sym]; ok {
			output.Sym = newIdx
			r.Outputs[name] = output
		}
	}

	// update symslots (highest used slot + 1)
	r.Symslots = sr.nextIdx
}

func (sr *SymbolRemap) remapOp(op *nir.Op) {
	for i := range op.Args {
		sr.remapSymExpr(&op.Args[i])
	}
	if op.Out != nil {
		sr.remapSym(op.Out)
	}
}

func (sr *SymbolRemap) remapSym(sym *int) {
	sr.slotMap[*sym] = sr.nextIdx
	*sym = sr.nextIdx
	sr.nextIdx++
}

func (sr *SymbolRemap) remapSymExpr(arg *nir.RawExpr) {
	if arg == nil || arg.Expr == nil {
		return
	}

	switch e := arg.Expr.(type) {
	case *nir.SymExpr:
		if newIdx, ok := sr.slotMap[e.Index]; ok {
			e.Index = newIdx
		}
	case *nir.IdxExpr:
		sr.remapSymExpr(&e.From)
		sr.remapSymExpr(&e.Index)
	case *nir.SelExpr:
		sr.remapSymExpr(&e.From)
		for i := range e.Exprs {
			sr.remapSymExpr(&e.Exprs[i])
		}
	case *nir.ArrExpr:
		for i := range e.Value {
			sr.remapSymExpr(&e.Value[i])
		}
	case *nir.ObjExpr:
		for _, prop := range e.Value {
			sr.remapSymExpr(&prop)
		}
	case *nir.SignalExpr:
		for i := range e.Args {
			sr.remapSymExpr(&e.Args[i])
		}
		if e.When != nil {
			sr.remapSymExpr(e.When)
		}
	case *nir.CallExpr:
		for i := range e.Args {
			sr.remapSymExpr(&e.Args[i])
		}
	case *nir.LambdaExpr:
		for p := range e.Params {
			sr.remapSym(&e.Params[p].SymIndex)
		}
		for i := range e.Ops {
			sr.remapOp(&e.Ops[i])
		}
	}
}
