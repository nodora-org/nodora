package crypto

import (
	"crypto/hmac"
	_sha256 "crypto/sha256"
	"encoding/hex"
	"fmt"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func hmacSha256() types.Func {
	return types.Func{
		Name: "hmac_sha256",
		Args: []types.ArgSpec{
			{Name: "key", Type: types.StringType, Required: true},
			{Name: "msg", Type: types.StringType, Required: true},
		},
		ReturnType: types.StringType,
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
		return core.U(), fmt.Errorf("expected string for 'key' argument, got %v", keyv.Type())
	}
	msg, ok := msgv.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'msg' argument, got %v", msgv.Type())
	}
	mac := hmac.New(_sha256.New, []byte(key))
	_, err := mac.Write([]byte(msg))
	if err != nil {
		return core.U(), err
	}

	return core.V(hex.EncodeToString(mac.Sum(nil))), nil
}
