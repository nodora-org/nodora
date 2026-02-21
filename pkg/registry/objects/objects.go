package objects

import (
	"nodora.org/nodora/pkg/registry"
)

func init() {
	registry.Global().Register("objects", get, keys, values, remove, filter, union, isSubset)
}
