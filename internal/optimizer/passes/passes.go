package passes

import "nodora.org/nodora/pkg/nir"

type Pass interface {
	Run(p *nir.Program) (changed bool, err error)
}
