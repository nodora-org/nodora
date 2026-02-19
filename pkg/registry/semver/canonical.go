package semver

import (
	"fmt"

	"golang.org/x/mod/semver"
	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func canonical() types.Func {
	return types.Func{
		Name:       "canonical",
		Args:       []types.ArgSpec{{Name: "version", Type: types.StringType, Required: true}},
		ReturnType: types.NumberType,
		Fn:         canonicalImpl,
	}
}

func canonicalImpl(args []core.Value) (core.Value, error) {
	version, ok := args[0].Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("invalid version type")
	}
	return core.V(semver.Canonical(version)), nil
}
