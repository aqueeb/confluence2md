//go:build linux && amd64

package pandoc

import _ "embed"

//go:embed bin/pandoc-linux-amd64
var embeddedBinary []byte
