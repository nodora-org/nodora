package semver

import (
	"fmt"

	"golang.org/x/mod/semver"
	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func isValid() types.Func {
	return types.Func{
		Name:       "is_valid",
		Args:       []types.ArgSpec{{Name: "version", Type: types.StringType, Required: true}},
		ReturnType: types.BoolType,
		Fn:         isValidImpl,
	}
}

func isValidImpl(args []core.Value) (core.Value, error) {
	version := args[0]
	v, ok := version.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'version' argument, got %v", version.Type())
	}
	return core.V(semver.IsValid(v)), nil
}
