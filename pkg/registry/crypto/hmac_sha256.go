package crypto

import (
	"crypto/hmac"
	_sha256 "crypto/sha256"
	"encoding/hex"
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/registry/types"
)

func hmacSha256() types.Func {
	return types.Func{
		Name: "hmac_sha256",
		Args: []types.ArgSpec{
			{Name: "key", Types: []string{"string"}, Required: true},
			{Name: "msg", Types: []string{"string"}, Required: true},
		},
		ReturnType: "string",
		Fn:         hmacSha256Impl,
	}
}

func hmacSha256Impl(args []core.Value) (core.Value, error) {
	keyv := args[0]
	msgv := args[1]
	if keyv.Undefined || msgv.Undefined {
		return core.U(), nil
	}
	key, ok := keyv.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("invalid key type")
	}
	msg, ok := msgv.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("invalid msg type")
	}
	mac := hmac.New(_sha256.New, []byte(key))
	_, err := mac.Write([]byte(msg))
	if err != nil {
		return core.U(), err
	}

	return core.V(hex.EncodeToString(mac.Sum(nil))), nil
}
