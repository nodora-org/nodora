package uuid

import (
	"fmt"

	"github.com/google/uuid"
	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func isValid() types.Func {
	return types.Func{
		Name:        "is_valid",
		Description: "Returns true if the string is a valid UUID.",
		Args:        []types.ArgSpec{{Name: "str", Description: "The string to validate.", Type: types.StringType, Required: true}},
		ReturnType:  types.BoolType,
		Fn:         isValidImpl,
	}
}

func isValidImpl(args []core.Value) (core.Value, error) {
	str := args[0]
	strVal, ok := str.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'str' argument, got %v", str.Type())
	}
	_, err := uuid.Parse(strVal)
	return core.V(err == nil), nil
}
