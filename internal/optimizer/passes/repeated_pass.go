package passes

import (
	"nodora.org/nodora/pkg/nir"
)

type RepeatedPass struct {
	passes   []Pass
	maxIters int
}

func NewRepeatedPass(maxIters int, passes ...Pass) *RepeatedPass {
	return &RepeatedPass{passes: passes, maxIters: maxIters}
}

func (rp *RepeatedPass) Run(p *nir.Ruleset) (bool, error) {
	everChanged := false

	for range rp.maxIters {
		changed := false

		for _, pass := range rp.passes {
			c, err := pass.Run(p)
			if err != nil {
				return everChanged, err
			}
			changed = changed || c
		}

		if !changed {
			break
		}
		everChanged = true
	}

	return everChanged, nil
}
