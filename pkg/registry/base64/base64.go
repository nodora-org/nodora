package base64

import (
	"nodora.org/nodora/pkg/registry"
)

func init() {
	registry.Register("base64", encode, decode, isValid)
}
