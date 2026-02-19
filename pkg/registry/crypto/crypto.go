package crypto

import (
	"nodora.org/nodora/pkg/registry"
)

func init() {
	registry.Register("crypto", sha1, sha256, hmacSha256)
}
