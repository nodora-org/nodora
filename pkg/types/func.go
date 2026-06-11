package types

import (
	"fmt"
	"strings"

	"nodora.org/nodora/pkg/core"
)

type ArgSpec struct {
	Name        string
	Description string
	Type        Type
	Required    bool
}

type Func struct {
	Namespace   string
	Name        string
	Args        []ArgSpec
	Description string
	ReturnType  Type
	Fn          func([]core.Value) (core.Value, error)
	Pure        bool
}

func (f *Func) RequiredArgCount() (n int) {
	for _, a := range f.Args {
		if a.Required {
			n++
		}
	}
	return
}

type Namespace struct {
	Name  string
	Funcs map[string]Func
}

func (fn *Func) FullPath() string {
	if fn.Namespace == "" {
		return fn.Name
	}
	return fn.Namespace + "::" + fn.Name
}

func (fn *Func) Signature() string {
	var args []string
	for _, arg := range fn.Args {
		suffix := ""
		if !arg.Required {
			suffix = "?"
		}
		argType := ": " + arg.Type.String()
		args = append(args, arg.Name+argType+suffix)
	}
	return fmt.Sprintf("%s(%s) -> %s", fn.FullPath(), strings.Join(args, ", "), fn.ReturnType.String())
}

func (fn *Func) CompactSignature() string {
	var args []string
	for _, arg := range fn.Args {
		suffix := ""
		if !arg.Required {
			suffix = "?"
		}
		args = append(args, arg.Name+suffix)
	}
	return fmt.Sprintf("%s(%s)", fn.FullPath(), strings.Join(args, ", "))
}
