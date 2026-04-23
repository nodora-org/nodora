package time

import (
	gotime "time"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func now() types.Func {
	return types.Func{
		Name:        "now",
		Description: "Returns the current timestamp in milliseconds.",
		Args:        []types.ArgSpec{},
		ReturnType:  types.NumberType,
		Fn:          nowImpl,
	}
}

func nowImpl(args []core.Value) (core.Value, error) {
	return core.V(float64(gotime.Now().UnixMilli())), nil
}
