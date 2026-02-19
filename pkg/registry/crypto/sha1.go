package crypto

import (
	_sha1 "crypto/sha1"
	"encoding/hex"
	"fmt"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func sha1() types.Func {
	return types.Func{
		Name:       "sha1",
		Args:       []types.ArgSpec{{Name: "str", Type: types.StringType, Required: true}},
		ReturnType: types.StringType,
		Fn:         sha1Impl,
	}
}

func sha1Impl(args []core.Value) (core.Value, error) {
	data, ok := args[0].Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("invalid str type")
	}
	hash := _sha1.Sum([]byte(data))
	return core.V(hex.EncodeToString(hash[:])), nil
}
