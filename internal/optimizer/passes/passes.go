package passes

import "nodora.org/nodora/pkg/nir"

type Pass interface {
	Run(p *nir.Ruleset) (changed bool, err error)
}
