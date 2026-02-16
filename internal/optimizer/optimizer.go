package optimizer

import (
	"nodora.org/nodora/pkg/nir"
)

type Pass interface {
	Run(p *nir.Program) error
}

type Optimizer struct {
	passes []Pass
}

func NewOptimizer() *Optimizer {
	return &Optimizer{
		passes: make([]Pass, 0),
	}
}

func (o *Optimizer) AddPass(p Pass) {
	o.passes = append(o.passes, p)
}

func (o *Optimizer) Run(p *nir.Program) error {
	for _, pass := range o.passes {
		if err := pass.Run(p); err != nil {
			return err
		}
	}
	return nil
}
