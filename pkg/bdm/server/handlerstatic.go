package server

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static
var staticFs embed.FS

// Use this during development to allow UI development without rebuilding.
// It will server the static UI file from disk instead from the embedded FS.
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
