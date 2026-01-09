//go:build windows && amd64

package pandoc

import _ "embed"

//go:embed bin/pandoc-windows-amd64.exe
var embeddedBinary []byte
