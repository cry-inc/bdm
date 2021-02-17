package server

import (
	_ "embed" // required for file embedding
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func createFaviconHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "image/x-icon")
		w.Write(icon)
	}
}

//go:embed static/favicon.ico
var icon []byte
