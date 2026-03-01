package ui

import (
	"embed"
	"io/fs"
)

//go:embed dist templates
var files embed.FS

// FS returns the root embedded filesystem containing both dist and templates
func FS() fs.FS {
	return files
}
