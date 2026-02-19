package strings

import (
	"nodora.org/nodora/pkg/registry"
)

func init() {
	registry.Global().Register("strings",
		isAlpha,
		isAlphanumeric,
		concat,
		contains,
		startsWith,
		endsWith,
		lower,
		upper,
		trim,
		trimStart,
		trimEnd,
		split,
		replace,
	)
}
