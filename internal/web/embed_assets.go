package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed assets
var embeddedAssets embed.FS

func assetFS() http.FileSystem {
	sub, err := fs.Sub(embeddedAssets, "assets")
	if err != nil {
		return http.FS(embeddedAssets)
	}
	return http.FS(sub)
}
