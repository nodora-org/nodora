package json

import (
	"nodora.org/nodora/pkg/registry"
)

func init() {
	registry.Register("json", isValid, marshal, unmarshal)
}
