package types

import (
	"fmt"
	"strings"
)

type TypeKind int

const (
	StringKind TypeKind = iota
	NumberKind
	BoolKind
	ObjectKind
	AnyKind
	LambdaKind
	UnknownKind
)

var typeKindNames = map[TypeKind]string{
	StringKind:  "string",
	NumberKind:  "number",
	BoolKind:    "bool",
	ObjectKind:  "object",
	AnyKind:     "any",
	LambdaKind:  "lambda",
	UnknownKind: "?",
}

type Type interface {
	String() string
	Equals(other Type) bool
	IsAssignableFrom(other Type) bool
	Kind() TypeKind
}

type SimpleType struct {
	kind TypeKind
}

func (t *SimpleType) String() string {
	if name, ok := typeKindNames[t.kind]; ok {
		return name
	}
	return typeKindNames[UnknownKind]
}

func (t *SimpleType) Equals(other Type) bool {
	if other == nil {
		return false
	}
	if simple, ok := other.(*SimpleType); ok {
		return t.kind == simple.kind
	}
	return false
}

func (t *SimpleType) IsAssignableFrom(other Type) bool {
	if other == nil {
		return false
	}
	if simple, ok := other.(*SimpleType); ok && simple.kind == UnknownKind {
		return true
	}
	if t.kind == AnyKind {
		return true
	}
	return t.Equals(other)
}

func (t *SimpleType) Kind() TypeKind {
	return t.kind
}

var (
	StringType  = &SimpleType{kind: StringKind}
	NumberType  = &SimpleType{kind: NumberKind}
	BoolType    = &SimpleType{kind: BoolKind}
	ObjectType  = &SimpleType{kind: ObjectKind}
	AnyType     = &SimpleType{kind: AnyKind}
	UnknownType = &SimpleType{kind: UnknownKind}
)

type ArrayType struct {
	Element Type
}

func (t *ArrayType) String() string {
	if t.Element == nil {
		return fmt.Sprintf("array<%s>", typeKindNames[UnknownKind])
	}
	return fmt.Sprintf("array<%s>", t.Element.String())
}

func (t *ArrayType) Equals(other Type) bool {
	if other == nil {
		return false
	}
	if arr, ok := other.(*ArrayType); ok {
		return t.Element.Equals(arr.Element)
	}
	return false
}

func (t *ArrayType) IsAssignableFrom(other Type) bool {
	if other == nil {
		return false
	}
	if simple, ok := other.(*SimpleType); ok && simple.kind == UnknownKind {
		return true
	}
	if arr, ok := other.(*ArrayType); ok {
		return t.Element.IsAssignableFrom(arr.Element)
	}
	return false
}

func (t *ArrayType) Kind() TypeKind {
	return AnyKind
}

func NewArrayType(element Type) *ArrayType {
	return &ArrayType{Element: element}
}

func IsArray(t Type) bool {
	_, ok := t.(*ArrayType)
	return ok
}

func IsSimpleType(t Type) bool {
	_, ok := t.(*SimpleType)
	return ok
}

func GetArrayElement(t Type) Type {
	if arr, ok := t.(*ArrayType); ok {
		return arr.Element
	}
	return nil
}

type LambdaType struct {
	Params     []Type
	ReturnType Type
}

func (t *LambdaType) String() string {
	params := make([]string, len(t.Params))
	for i, p := range t.Params {
		params[i] = p.String()
	}
	ret := "?"
	if t.ReturnType != nil {
		ret = t.ReturnType.String()
	}
	return fmt.Sprintf("|%s| -> %s", strings.Join(params, ", "), ret)
}

func (t *LambdaType) Equals(other Type) bool {
	if other == nil {
		return false
	}
	if ol, ok := other.(*LambdaType); ok {
		if len(t.Params) != len(ol.Params) {
			return false
		}
		for i, p := range t.Params {
			if !p.Equals(ol.Params[i]) {
				return false
			}
		}
		if t.ReturnType == nil && ol.ReturnType == nil {
			return true
		}
		if t.ReturnType == nil || ol.ReturnType == nil {
			return false
		}
		return t.ReturnType.Equals(ol.ReturnType)
	}
	return false
}

func (t *LambdaType) IsAssignableFrom(other Type) bool {
	if other == nil {
		return false
	}

	o, ok := other.(*LambdaType)
	if !ok {
		return false
	}

	if len(t.Params) != len(o.Params) {
		return false
	}

	for i := range t.Params {
		if !t.Params[i].IsAssignableFrom(o.Params[i]) {
			return false
		}
	}

	if t.ReturnType == nil {
		return o.ReturnType == nil
	}

	return t.ReturnType.IsAssignableFrom(o.ReturnType)
}

func (t *LambdaType) Kind() TypeKind {
	return LambdaKind
}

func NewLambdaType(params []Type, retType Type) *LambdaType {
	return &LambdaType{
		Params:     params,
		ReturnType: retType,
	}
}

type UnionType struct {
	Types []Type
}

func (t *UnionType) String() string {
	parts := make([]string, len(t.Types))
	for i, ty := range t.Types {
		parts[i] = ty.String()
	}
	return "union<" + strings.Join(parts, "|") + ">"
}

func (t *UnionType) Equals(other Type) bool {
	if other == nil {
		return false
	}
	ou, ok := other.(*UnionType)
	if !ok {
		return false
	}
	if len(t.Types) != len(ou.Types) {
		return false
	}
	for i, ty := range t.Types {
		if !ty.Equals(ou.Types[i]) {
			return false
		}
	}
	return true
}

func (t *UnionType) IsAssignableFrom(other Type) bool {
	if other == nil {
		return false
	}
	if simple, ok := other.(*SimpleType); ok && simple.kind == UnknownKind {
		return true
	}
	for _, ty := range t.Types {
		if ty.IsAssignableFrom(other) {
			return true
		}
	}
	return false
}

func (t *UnionType) Kind() TypeKind {
	return AnyKind
}

func NewUnionType(types ...Type) *UnionType {
	return &UnionType{Types: types}
}

func IsUnion(t Type) bool {
	_, ok := t.(*UnionType)
	return ok
}
