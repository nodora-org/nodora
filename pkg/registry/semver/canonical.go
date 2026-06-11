package semver

import (
	"fmt"

	"golang.org/x/mod/semver"
	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func canonical() types.Func {
	return types.Func{
		Name:        "canonical",
		Description: "Returns the canonical form of a semantic version string.",
		Args:        []types.ArgSpec{{Name: "version", Description: "The version string to canonicalize.", Type: types.StringType, Required: true}},
		ReturnType:  types.NumberType,
		Fn:          canonicalImpl,
		Pure:        true,
	}
}

func canonicalImpl(args []core.Value) (core.Value, error) {
	version := args[0]
	v, ok := version.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'version' argument, got %v", version.Type())
	}
	return core.V(semver.Canonical(v)), nil
}
