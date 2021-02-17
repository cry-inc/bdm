package server

import (
	"embed"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

//go:embed static/*
var staticFs embed.FS

func createHTMLHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		html, err := staticFs.ReadFile("static/index.html")
		if err != nil {
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(html)
	}
}

func createFaviconHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		icon, err := staticFs.ReadFile("static/favicon.ico")
		if err != nil {
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/x-icon")
		w.Write(icon)
	}
}
