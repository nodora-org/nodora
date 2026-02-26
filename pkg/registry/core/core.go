package core

import "nodora.org/nodora/internal/types"

func GetFuncs() []func() types.Func {
	return []func() types.Func{
		int_,
		length,
		isDefined,
		isEmpty,
		isNumber,
		isString,
		isBool,
		isArray,
		isObject,
		sprintf,
		some,
		every,
		min,
		max,
		sum,
		product,
	}
}
