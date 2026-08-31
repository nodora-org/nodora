package core

import (
	"encoding/binary"
	"fmt"
	"math"
	"sort"
)

// canonical byte-encoding tags
const (
	tagUndefined byte = 0x00
	tagNull      byte = 0x01
	tagFalse     byte = 0x02
	tagTrue      byte = 0x03
	tagNumber    byte = 0x04
	tagString    byte = 0x05
	tagArray     byte = 0x06
	tagObject    byte = 0x07
)

const canonicalNaN uint64 = 0x7FF8000000000000

// Returns a canonical, injective binary encoding of Value.
func (v Value) ToCanonicalBytes() ([]byte, error) {
	return appendCanonical(nil, v)
}

func appendCanonical(buf []byte, v Value) ([]byte, error) {
	if v.IsUndefined() {
		return append(buf, tagUndefined), nil
	}

	switch raw := v.Raw.(type) {
	case Value:
		return appendCanonical(buf, raw)
	case nil:
		return append(buf, tagNull), nil
	case bool:
		if raw {
			return append(buf, tagTrue), nil
		}
		return append(buf, tagFalse), nil
	case float64:
		bits := math.Float64bits(raw)
		switch {
		case raw == 0: // collapse -0 into +0
			bits = 0
		case math.IsNaN(raw):
			bits = canonicalNaN
		}
		buf = append(buf, tagNumber)
		return binary.BigEndian.AppendUint64(buf, bits), nil
	case string:
		buf = append(buf, tagString)
		buf = binary.AppendUvarint(buf, uint64(len(raw)))
		return append(buf, raw...), nil
	case []Value:
		buf = append(buf, tagArray)
		buf = binary.AppendUvarint(buf, uint64(len(raw)))
		var err error
		for i := range raw {
			if buf, err = appendCanonical(buf, raw[i]); err != nil {
				return nil, err
			}
		}
		return buf, nil
	case ValueMap:
		buf = append(buf, tagObject)
		buf = binary.AppendUvarint(buf, uint64(len(raw)))
		keys := make([]string, 0, len(raw))
		for k := range raw {
			keys = append(keys, k)
		}
		sort.Strings(keys) // byte-order sort => order-independent encoding
		var err error
		for _, k := range keys {
			buf = binary.AppendUvarint(buf, uint64(len(k)))
			buf = append(buf, k...)
			if buf, err = appendCanonical(buf, raw[k]); err != nil {
				return nil, err
			}
		}
		return buf, nil
	default:
		return nil, fmt.Errorf("cannot encode value of type %s", v.Type())
	}
}
