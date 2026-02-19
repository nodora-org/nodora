package math

import (
	"nodora.org/nodora/pkg/registry"
)

func init() {
	registry.Global().Register("math", abs, round, ceil, floor, maxFunc, minFunc, clamp)
}
