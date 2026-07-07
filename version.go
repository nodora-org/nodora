package version

import (
	"runtime/debug"
	"sync"
)

const modulePath = "nodora.org/nodora"

// optional override injected at release-build time via ldflags:
// -ldflags "-X nodora.org/nodora.version=v1.2.3"
var version string

var Get = sync.OnceValue(func() string {
	if version != "" {
		return version
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		v := info.Main.Version
		if info.Main.Path == modulePath && v != "" && v != "(devel)" {
			return v
		}
		for _, dep := range info.Deps {
			if dep.Path == modulePath {
				if dep.Replace != nil && dep.Replace.Version != "" {
					return dep.Replace.Version
				}
				if dep.Version != "" {
					return dep.Version
				}
			}
		}
	}

	return "dev-build"
})
