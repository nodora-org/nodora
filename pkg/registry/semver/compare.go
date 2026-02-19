package semver

import (
	"fmt"

	"golang.org/x/mod/semver"
	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/registry/types"
)

func compare() types.Func {
	return types.Func{
		Name: "compare",
		Args: []types.ArgSpec{
			{Name: "v", Types: []string{"string"}, Required: true},
			{Name: "w", Types: []string{"string"}, Required: true},
		},
		ReturnType: "number",
		Fn:         compareImpl,
	}
}

func compareImpl(args []core.Value) (core.Value, error) {
	v, ok := args[0].Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("invalid v type")
	}
	w, ok := args[1].Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("invalid w type")
	}
	return core.V(semver.Compare(v, w)), nil
}
