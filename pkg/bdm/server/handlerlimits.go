package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/cry-inc/bdm/pkg/bdm"
)

func createLimitsHandler(limits *bdm.ManifestLimits, tokens Tokens) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		if !hasReadToken(req, tokens) {
			http.Error(writer, "Invalid token", http.StatusUnauthorized)
			return
		}

		jsonData, err := json.Marshal(limits)
		if err != nil {
			log.Print(fmt.Errorf("error marshalling limits JSON: %w", err))
			http.Error(writer, "Failed to generate JSON data", http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(jsonData)
	}
}
