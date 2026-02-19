package types

import (
	"fmt"
	"strings"

	"nodora.org/nodora/pkg/core"
)

type ArgSpec struct {
	Name     string
	Types    []string
	Required bool
}

type Func struct {
	Namespace  string
	Name       string
	Args       []ArgSpec
	ReturnType string
	Fn         func([]core.Value) (core.Value, error)
}

type Namespace struct {
	Name  string
	Funcs map[string]Func
}

type Registry struct {
	ns map[string]*Namespace
}

func NewRegistry() Registry {
	return Registry{
		ns: make(map[string]*Namespace),
	}
}

func (r *Registry) Exists(namespace, name string) bool {
	if ns, ok := r.ns[namespace]; ok {
		_, exists := ns.Funcs[name]
		return exists
	}
	return false
}

func (r *Registry) Get(namespace, name string) (*Func, bool) {
	if ns, ok := r.ns[namespace]; ok {
		fn, exists := ns.Funcs[name]
		return &fn, exists
	}
	return nil, false
}

func (r *Registry) Register(namespace string, generators ...func() Func) error {
	if _, exists := r.ns[namespace]; exists {
		return fmt.Errorf("namespace '%s' already exists", namespace)
	}
	r.ns[namespace] = &Namespace{Name: namespace, Funcs: map[string]Func{}}
	for _, gen := range generators {
		f := gen()
		if r.Exists(namespace, f.Name) {
			return fmt.Errorf("function '%s' already registered in namespace '%s'", f.Name, namespace)
		}
		f.Namespace = namespace
		r.ns[namespace].Funcs[f.Name] = f
	}
	return nil
}

func (fn *Func) FullPath() string {
	return fn.Namespace + "::" + fn.Name
}

func (fn *Func) Signature() string {
	var args []string
	for _, arg := range fn.Args {
		suffix := ""
		if !arg.Required {
			suffix = "?"
		}
		types := ""
		if len(arg.Types) > 0 {
			types = ": " + strings.Join(arg.Types, "|")
		}
		args = append(args, arg.Name+types+suffix)
	}
	return fmt.Sprintf("%s(%s) -> %s", fn.FullPath(), strings.Join(args, ", "), fn.ReturnType)
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
