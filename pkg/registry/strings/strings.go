package strings

import (
	"nodora.org/nodora/pkg/registry"
)

func init() {
	registry.Register("strings",
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
