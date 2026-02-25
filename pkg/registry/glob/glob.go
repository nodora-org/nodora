package glob

import "nodora.org/nodora/pkg/registry"

func init() {
	registry.Global().Register("glob", match)
}
