package uuid

import (
	"fmt"

	"github.com/google/uuid"
	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func isValid() types.Func {
	return types.Func{
		Name:       "is_valid",
		Args:       []types.ArgSpec{{Name: "str", Type: types.StringType, Required: true}},
		ReturnType: types.BoolType,
		Fn:         isValidImpl,
	}
}

func isValidImpl(args []core.Value) (core.Value, error) {
	str, ok := args[0].Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("invalid str type")
	}
	_, err := uuid.Parse(str)
	return core.V(err == nil), nil
}
