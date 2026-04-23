package crypto

import (
	_sha256 "crypto/sha256"
	"encoding/hex"
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func sha256() types.Func {
	return types.Func{
		Name:        "sha256",
		Description: "Computes the SHA-256 hash of a string, returned as a hex string.",
		Args:        []types.ArgSpec{{Name: "str", Description: "The string to hash.", Type: types.StringType, Required: true}},
		ReturnType:  types.StringType,
		Fn:          sha256Impl,
	}
}

func sha256Impl(args []core.Value) (core.Value, error) {
	str := args[0]
	data, ok := str.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'str' argument, got %v", str.Type())
	}
	hash := _sha256.Sum256([]byte(data))
	return core.V(hex.EncodeToString(hash[:])), nil
}
