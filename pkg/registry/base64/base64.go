package base64

import (
	"nodora.org/nodora/pkg/registry"
)

func init() {
	registry.Global().Register("base64", encode, decode, isValid)
}
