package semver

import (
	"fmt"

	"golang.org/x/mod/semver"
	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/registry/types"
)

func canonical() types.Func {
	return types.Func{
		Name:       "canonical",
		Args:       []types.ArgSpec{{Name: "version", Types: []string{"string"}, Required: true}},
		ReturnType: "number",
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
