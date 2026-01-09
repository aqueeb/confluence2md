//go:build darwin && amd64

package pandoc

import _ "embed"

//go:embed bin/pandoc-darwin-amd64
var embeddedBinary []byte
