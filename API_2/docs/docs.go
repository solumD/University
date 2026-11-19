package docs

import "embed"

//go:embed openapi.yaml index.html
var Files embed.FS
