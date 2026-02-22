package all

// Importing this package ensures that all builtins are registered in the registry (init() registration)
// Usage: import _ "nodora.org/nodora/pkg/registry/all"

import (
	_ "nodora.org/nodora/pkg/registry/arrays"
	_ "nodora.org/nodora/pkg/registry/base64"
	_ "nodora.org/nodora/pkg/registry/base64url"
	_ "nodora.org/nodora/pkg/registry/conv"
	_ "nodora.org/nodora/pkg/registry/crypto"
	_ "nodora.org/nodora/pkg/registry/json"
	_ "nodora.org/nodora/pkg/registry/math"
	_ "nodora.org/nodora/pkg/registry/objects"
	_ "nodora.org/nodora/pkg/registry/semver"
	_ "nodora.org/nodora/pkg/registry/strings"
	_ "nodora.org/nodora/pkg/registry/uuid"
)
