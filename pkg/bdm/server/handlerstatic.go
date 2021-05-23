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
	var fileSystem http.FileSystem = nil
	if developmentMode {
		fileSystem = http.Dir("pkg/bdm/server/static")
	} else {
		embeddedFileSystem, err := fs.Sub(staticFs, "static")
		if err != nil {
			panic(err)
		}
		fileSystem = http.FS(embeddedFileSystem)
	}
	return http.FileServer(fileSystem).ServeHTTP
}
