package optimizer

import (
	"nodora.org/nodora/internal/optimizer/passes"
	"nodora.org/nodora/pkg/nir"
)

type Optimizer struct {
	passes []passes.Pass
}

func NewOptimizer() *Optimizer {
	return &Optimizer{
		passes: make([]passes.Pass, 0),
	}
}

func (o *Optimizer) AddPass(p passes.Pass) {
	o.passes = append(o.passes, p)
}

func (o *Optimizer) Run(p *nir.Ruleset) error {
	for _, pass := range o.passes {
		if _, err := pass.Run(p); err != nil {
			return err
		}
	}
	return nil
}
