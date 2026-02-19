package types

import (
	"fmt"

	"nodora.org/nodora/internal/types"
)

type Registry struct {
	ns map[string]*types.Namespace
}

func NewRegistry() Registry {
	return Registry{
		ns: make(map[string]*types.Namespace),
	}
}

func (r *Registry) Exists(namespace, name string) bool {
	if ns, ok := r.ns[namespace]; ok {
		_, exists := ns.Funcs[name]
		return exists
	}
	return false
}

func (r *Registry) Get(namespace, name string) (*types.Func, bool) {
	if ns, ok := r.ns[namespace]; ok {
		fn, exists := ns.Funcs[name]
		return &fn, exists
	}
	return nil, false
}

func (r *Registry) Register(namespace string, generators ...func() types.Func) error {
	if _, exists := r.ns[namespace]; exists {
		return fmt.Errorf("namespace '%s' already exists", namespace)
	}
	r.ns[namespace] = &types.Namespace{Name: namespace, Funcs: map[string]types.Func{}}
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
