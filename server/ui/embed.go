package ui

import (
	"embed"
	"io/fs"
)

//go:embed dist templates
var files embed.FS

func FS() fs.FS {
	sub, err := fs.Sub(files, "dist")
	if err != nil {
		panic(err)
	}
	return sub
}
