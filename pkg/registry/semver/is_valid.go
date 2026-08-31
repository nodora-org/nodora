package semver

import (
	"fmt"

	"golang.org/x/mod/semver"
	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func isValid() types.Func {
	return types.Func{
		Name:        "is_valid",
		Description: "Returns true if the string is a valid semantic version.",
		Args:        []types.ArgSpec{{Name: "version", Description: "The version string to validate.", Type: types.StringType, Required: true}},
		ReturnType:  types.BoolType,
		Fn:          isValidImpl,
		Pure:        true,
	}
}

func isValidImpl(args []core.Value) (core.Value, error) {
	version := args[0]
	v, ok := version.AsString()
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'version' argument, got %v", version.Type())
	}
	return core.V(semver.IsValid(v)), nil
}
