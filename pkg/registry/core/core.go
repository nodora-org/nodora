package core

import "nodora.org/nodora/pkg/registry/types"

func GetFuncs() []func() types.Func {
	return []func() types.Func{
		isDefined, length, isEmpty, sprintf,
	}
}
