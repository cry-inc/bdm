package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/cry-inc/bdm/pkg/bdm"
	"github.com/cry-inc/bdm/pkg/bdm/store"
)

func createCheckObjectsHandler(packageStore store.Store, tokens Tokens) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		if !hasReadToken(req, tokens) {
			http.Error(writer, "Invalid token", http.StatusUnauthorized)
			return
		}

		objects, err := checkStoreForObjects(req.Body, packageStore)
		if err != nil {
			http.Error(writer, "Bad request", http.StatusBadRequest)
			return
		}

		json, err := json.Marshal(objects)
		if err != nil {
			log.Print(fmt.Errorf("error marshalling objects to JSON: %w", err))
			http.Error(writer, "Failed to generate JSON data", http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(json)
	}
}

func createUploadObjectsHandler(packageStore store.Store, limits *bdm.ManifestLimits, tokens Tokens) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		if !hasWriteToken(req, tokens) {
			http.Error(writer, "Invalid token", http.StatusUnauthorized)
			return
		}

		objects, err := streamObjectsToStore(req.Body, packageStore, limits.MaxFileSize)
		if err != nil {
			http.Error(writer, "Bad request", http.StatusBadRequest)
			return
		}

		json, err := json.Marshal(objects)
		if err != nil {
			log.Print(fmt.Errorf("error marshalling objects to JSON: %w", err))
			http.Error(writer, "Failed to generate JSON data", http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(json)
	}
}

func createDownloadObjectsHandler(packageStore store.Store, tokens Tokens) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		if !hasReadToken(req, tokens) {
			http.Error(writer, "Invalid token", http.StatusUnauthorized)
			return
		}

		err := streamObjectsFromStore(req.Body, packageStore, writer)
		if err != nil {
			http.Error(writer, "Bad request", http.StatusBadRequest)
			return
		}
	}
}
