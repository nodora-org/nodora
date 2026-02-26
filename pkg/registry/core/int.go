package core

import (
	"fmt"
	"math"
	"strconv"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func int_() types.Func {
	return types.Func{
		Name:        "int",
		Description: "Returns the integer value of x.",
		Args: []types.ArgSpec{{
			Name:        "x",
			Description: "The value to convert.",
			Type:        types.AnyType,
			Required:    true,
		}},
		ReturnType: types.NumberType,
		Fn:         intImpl,
	}
}

func intImpl(args []core.Value) (core.Value, error) {
	value := args[0]
	if value.Undefined {
		return core.U(), nil
	}
	switch v := value.Raw.(type) {
	case float64:
		return core.V(math.Trunc(v)), nil
	case bool:
		if v {
			return core.V(float64(1)), nil
		}
		return core.V(float64(0)), nil
	case string:
		val, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return core.U(), fmt.Errorf("cannot parse %q as number", v)
		}
		return core.V(math.Trunc(val)), nil
	default:
		return core.U(), fmt.Errorf("%v is not convertible to int", value.Type())
	}
}
