package core

import "nodora.org/nodora/internal/types"

func GetFuncs() []func() types.Func {
	return []func() types.Func{
		isDefined, length, isEmpty, sprintf,
	}
}
