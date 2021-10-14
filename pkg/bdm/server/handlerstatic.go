package server

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static
var staticFs embed.FS

// Use this during development to test web UI changes without rebuilding.
// It will serve the static UI assets from disk instead from the embedded FS.
const developmentMode = false

func createStaticHandler() http.HandlerFunc {
	if developmentMode {
		return http.FileServer(http.Dir("pkg/bdm/server/static")).ServeHTTP
	}

	embeddedFileSystem, err := fs.Sub(staticFs, "static")
	if err != nil {
		panic(err)
	}

	return http.FileServer(http.FS(embeddedFileSystem)).ServeHTTP
}
