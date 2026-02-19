package nir

import (
	"fmt"

	"nodora.org/nodora/pkg/core"
)

type Float interface {
	~float32 | ~float64
}

type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type Numeric interface {
	Signed | Unsigned | Float
}

func contains(arr []core.Value, el core.Value) bool {
	for _, v := range arr {
		if v.ToRaw() == el.ToRaw() {
			return true
		}
	}
	return false
}

func validateArgs(o *Op, argCount int) error {
	if len(o.Args) != argCount {
		return fmt.Errorf("invalid operation arguments for '%s'", o.Kind)
	}
	return nil
}

func validateOp(o *Op, ctx *EvaluationContext, argCount int) error {
	if err := validateArgs(o, argCount); err != nil {
		return err
	}
	if o.Out == nil {
		return fmt.Errorf("output index required for '%s'", o.Kind)
	}
	if *o.Out < 0 || *o.Out >= len(ctx.Slots) {
		return fmt.Errorf("output index %d is out of bounds", *o.Out)
	}
	return nil
}
