package server

import (
	_ "embed" // required for file embedding
	"net/http"

	"github.com/julienschmidt/httprouter"
)

//go:embed static/index.html
var html []byte

//go:embed static/favicon.ico
var icon []byte

func createHTMLHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(html)
	}
}

func createFaviconHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "image/x-icon")
		w.Write(icon)
	}
}
