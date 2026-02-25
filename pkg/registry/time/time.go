package time

import (
	"nodora.org/nodora/pkg/registry"
)

func init() {
	registry.Global().Register("time", now, add, sub, addUnits, subUnits, parseRFC3339, formatRFC3339, startOf, endOf)
}
