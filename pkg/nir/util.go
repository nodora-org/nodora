package nir

import (
	"nodora.org/nodora/pkg/core"
)

func contains(arr []core.Value, el core.Value) bool {
	for _, v := range arr {
		if core.SafeEquals(v, el) {
			return true
		}
	}
	return false
}

