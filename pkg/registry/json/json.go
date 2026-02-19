package json

import (
	"nodora.org/nodora/pkg/registry"
)

func init() {
	registry.Global().Register("json", isValid, marshal, unmarshal)
}
