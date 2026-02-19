package registry

import (
	"nodora.org/nodora/pkg/registry/core"
	"nodora.org/nodora/pkg/registry/types"
)

var registry types.Registry

func init() {
	registry = types.NewRegistry()
	registry.Register("", core.GetFuncs()...)
}

func Exists(namespace, name string) bool {
	return registry.Exists(namespace, name)
}

func Get(namespace, name string) (*types.Func, bool) {
	return registry.Get(namespace, name)
}

func Register(namespace string, generators ...func() types.Func) error {
	return registry.Register(namespace, generators...)
}
