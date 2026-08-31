package core

import (
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func hash() types.Func {
	return types.Func{
		Name:        "hash",
		Description: "Computes the XXH32 hash for the given value.",
		Args: []types.ArgSpec{
			{
				Name:        "value",
				Description: "The value to hash.",
				Type:        types.AnyType,
				Required:    true,
			},
			{
				Name:        "seed",
				Description: "The seed for the hash function.",
				Type:        types.NumberType,
				Required:    false,
			},
		},
		ReturnType: types.NumberType,
		Fn:         hashImpl,
		Pure:       true,
	}
}

func hashImpl(args []core.Value) (core.Value, error) {
	var seed uint32 = 0
	if len(args) > 1 && !args[1].IsUndefined() {
		seedVal, ok := args[1].AsFloat()
		if !ok {
			return core.U(), fmt.Errorf("seed must be a number")
		}
		seed = uint32(seedVal)
	}

	value := args[0]
	data, err := value.ToCanonicalBytes()
	if err != nil {
		return core.U(), err
	}
	hash := xxhash32(data, seed)
	return core.Num(float64(hash)), nil
}

const (
	prime32_1 = 2654435761
	prime32_2 = 2246822519
	prime32_3 = 3266489917
	prime32_4 = 668265263
	prime32_5 = 374761393
)

func xxhash32(input []byte, seed uint32) uint32 {
	n := len(input)
	h32 := uint32(n)

	if n < 16 {
		h32 += seed + prime32_5
	} else {
		v1 := seed + prime32_1 + prime32_2
		v2 := seed + prime32_2
		v3 := seed
		v4 := seed - prime32_1
		p := 0
		for n := n - 16; p <= n; p += 16 {
			sub := input[p:][:16]
			v1 = rol13(v1+u32(sub[:])*prime32_2) * prime32_1
			v2 = rol13(v2+u32(sub[4:])*prime32_2) * prime32_1
			v3 = rol13(v3+u32(sub[8:])*prime32_2) * prime32_1
			v4 = rol13(v4+u32(sub[12:])*prime32_2) * prime32_1
		}
		input = input[p:]
		n -= p
		h32 += rol1(v1) + rol7(v2) + rol12(v3) + rol18(v4)
	}

	p := 0
	for n := n - 4; p <= n; p += 4 {
		h32 += u32(input[p:p+4]) * prime32_3
		h32 = rol17(h32) * prime32_4
	}
	for p < n {
		h32 += uint32(input[p]) * prime32_5
		h32 = rol11(h32) * prime32_1
		p++
	}

	h32 ^= h32 >> 15
	h32 *= prime32_2
	h32 ^= h32 >> 13
	h32 *= prime32_3
	h32 ^= h32 >> 16

	return h32
}

func u32(buf []byte) uint32 {
	return uint32(buf[0]) | uint32(buf[1])<<8 | uint32(buf[2])<<16 | uint32(buf[3])<<24
}

func rol1(u uint32) uint32 {
	return u<<1 | u>>31
}

func rol7(u uint32) uint32 {
	return u<<7 | u>>25
}

func rol11(u uint32) uint32 {
	return u<<11 | u>>21
}

func rol12(u uint32) uint32 {
	return u<<12 | u>>20
}

func rol13(u uint32) uint32 {
	return u<<13 | u>>19
}

func rol17(u uint32) uint32 {
	return u<<17 | u>>15
}

func rol18(u uint32) uint32 {
	return u<<18 | u>>14
}
