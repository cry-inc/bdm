package server

import (
	_ "embed" // required for file embedding
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func createHTMLHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(html)
	}
}

//go:embed static/index.html
var html []byte
