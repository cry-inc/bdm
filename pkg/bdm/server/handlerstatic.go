package server

import (
	"embed"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/julienschmidt/httprouter"
)

//go:embed static
var staticFs embed.FS

func createStaticHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}
		data, err := staticFs.ReadFile("static" + path)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		ext := filepath.Ext(path)
		if len(ext) > 0 {
			mimeType := mime.TypeByExtension("." + ext)
			if len(mimeType) > 0 {
				w.Header().Set("Content-Type", mimeType)
			}
		}
		w.Write(data)
	}
}