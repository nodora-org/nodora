package conv

import (
	"nodora.org/nodora/pkg/registry"
)

func init() {
	registry.Global().Register("conv", atoi, atof)
}
