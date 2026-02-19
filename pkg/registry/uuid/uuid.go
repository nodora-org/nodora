package uuid

import (
	"nodora.org/nodora/pkg/registry"
)

func init() {
	registry.Global().Register("uuid", isValid)
}
