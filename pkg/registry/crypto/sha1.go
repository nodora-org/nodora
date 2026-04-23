package crypto

import (
	_sha1 "crypto/sha1"
	"encoding/hex"
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func sha1() types.Func {
	return types.Func{
		Name:        "sha1",
		Description: "Computes the SHA-1 hash of a string, returned as a hex string.",
		Args:        []types.ArgSpec{{Name: "str", Description: "The string to hash.", Type: types.StringType, Required: true}},
		ReturnType:  types.StringType,
		Fn:          sha1Impl,
	}
}

func sha1Impl(args []core.Value) (core.Value, error) {
	str := args[0]
	data, ok := str.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'str' argument, got %v", str.Type())
	}
	hash := _sha1.Sum([]byte(data))
	return core.V(hex.EncodeToString(hash[:])), nil
}
