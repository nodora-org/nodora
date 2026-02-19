package semver

import (
	"nodora.org/nodora/pkg/registry"
)

func init() {
	registry.Register("semver", isValid, compare, canonical)
}
