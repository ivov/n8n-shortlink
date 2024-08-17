package internal

import (
	"embed"
	"io/fs"
)

//go:embed all:static
var embeddedFs embed.FS

// Static is an embedded-in-binary static files dir for the API to serve.
func Static() fs.FS {
	subtree, err := fs.Sub(embeddedFs, "static")
	if err != nil {
		panic(err)
	}

	return subtree
}
