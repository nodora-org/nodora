package uuid

import (
	"nodora.org/nodora/pkg/registry"
)

func init() {
	registry.Register("uuid", isValid)
}
