// Package web embeds the built Vue app. Run `npm run build` in this directory before `go build`.
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

func Dist() fs.FS {
	sub, _ := fs.Sub(dist, "dist")
	return sub
}
