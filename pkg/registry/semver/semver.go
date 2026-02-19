package semver

import (
	"nodora.org/nodora/pkg/registry"
)

func init() {
	registry.Global().Register("semver", isValid, compare, canonical)
}
