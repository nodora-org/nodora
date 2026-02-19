package math

import (
	"nodora.org/nodora/pkg/registry"
)

func init() {
	registry.Register("math", abs, round, ceil, floor, max, min, clamp)
}
