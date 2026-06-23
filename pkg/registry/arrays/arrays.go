package arrays

import (
	"nodora.org/nodora/pkg/registry"
)

func init() {
	registry.Global().Register("arrays", concat, flatten, reverse, slice, zip, groupBy, findIndex, mapFunc, filter, unique, reduce)
}
