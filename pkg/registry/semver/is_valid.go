package semver

import (
	"fmt"

	"golang.org/x/mod/semver"
	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/registry/types"
)

func isValid() types.Func {
	return types.Func{
		Name:       "is_valid",
		Args:       []types.ArgSpec{{Name: "version", Types: []string{"string"}, Required: true}},
		ReturnType: "bool",
		Fn:         isValidImpl,
	}
}

func isValidImpl(args []core.Value) (core.Value, error) {
	version, ok := args[0].Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("invalid version type")
	}
	return core.V(semver.IsValid(version)), nil
}
