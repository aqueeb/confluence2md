//go:build darwin && arm64

package pandoc

import _ "embed"

//go:embed bin/pandoc-darwin-arm64
var embeddedBinary []byte
