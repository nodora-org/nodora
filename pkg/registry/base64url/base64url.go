package base64url

import (
	"nodora.org/nodora/pkg/registry"
)

func init() {
	registry.Global().Register("base64url", encode, decode, isValid)
}
