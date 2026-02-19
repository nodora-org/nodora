package types

import "fmt"

type TypeKind int

const (
	StringKind TypeKind = iota
	NumberKind
	BoolKind
	ObjectKind
	AnyKind
	UnknownKind
)

var typeKindNames = map[TypeKind]string{
	StringKind:  "string",
	NumberKind:  "number",
	BoolKind:    "bool",
	ObjectKind:  "object",
	AnyKind:     "any",
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
	if otherSimple, ok := other.(*SimpleType); ok {
		return t.kind == otherSimple.kind
	}
	return false
}

func (t *SimpleType) IsAssignableFrom(other Type) bool {
	if other == nil {
		return false
	}
	// unknown is assignable to any type
	if otherUnknown, ok := other.(*SimpleType); ok && otherUnknown.kind == UnknownKind {
		return true
	}
	// any type can accept any value
	if t.kind == AnyKind {
		return true
	}
	// check exact match
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
	if otherArray, ok := other.(*ArrayType); ok {
		return t.Element.Equals(otherArray.Element)
	}
	return false
}

func (t *ArrayType) IsAssignableFrom(other Type) bool {
	if other == nil {
		return false
	}
	// unknown is assignable to any type
	if otherSimple, ok := other.(*SimpleType); ok && otherSimple.kind == UnknownKind {
		return true
	}
	if otherArray, ok := other.(*ArrayType); ok {
		return t.Element.IsAssignableFrom(otherArray.Element)
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
