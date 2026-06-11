package uuid

import (
	"encoding/hex"
	"fmt"
	"strings"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func isValid() types.Func {
	return types.Func{
		Name:        "is_valid",
		Description: "Returns true if the string is a valid UUID.",
		Args:        []types.ArgSpec{{Name: "str", Description: "The string to validate.", Type: types.StringType, Required: true}},
		ReturnType:  types.BoolType,
		Fn:          isValidImpl,
		Pure:        true,
	}
}

func isValidImpl(args []core.Value) (core.Value, error) {
	str := args[0]
	strVal, ok := str.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'str' argument, got %v", str.Type())
	}
	return core.V(isValidUUID(strVal)), nil
}

func isValidUUID(s string) bool {
	switch len(s) {
	case 36:
		// standard
	case 36 + 9:
		if !strings.EqualFold(s[:9], "urn:uuid:") {
			return false
		}
		s = s[9:]
	case 36 + 2:
		if s[0] != '{' || s[len(s)-1] != '}' {
			return false
		}
		s = s[1 : len(s)-1]
	case 32:
		_, err := hex.DecodeString(s)
		return err == nil
	default:
		return false
	}
	if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return false
	}
	stripped := s[:8] + s[9:13] + s[14:18] + s[19:23] + s[24:]
	_, err := hex.DecodeString(stripped)
	return err == nil
}
