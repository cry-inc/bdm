package server

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cry-inc/bdm/pkg/bdm"
	"github.com/cry-inc/bdm/pkg/bdm/store"
	"github.com/go-chi/chi/v5"
)

func createFilesHandler(packageStore store.Store, tokens Tokens) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		if !hasReadToken(req, tokens) {
			http.Error(writer, "Invalid token", http.StatusUnauthorized)
			return
		}

		name := chi.URLParam(req, "name")
		validName := bdm.ValidatePackageName(name)
		if !validName {
			http.Error(writer, "Bad package name", http.StatusBadRequest)
			return
		}

		versionString := chi.URLParam(req, "version")
		version, err := strconv.Atoi(versionString)
		if err != nil || version <= 0 {
			http.Error(writer, "Bad package version", http.StatusBadRequest)
			return
		}

		manifest, err := packageStore.GetManifest(name, uint(version))
		if err != nil || manifest == nil {
			http.Error(writer, "Package does not exist", http.StatusNotFound)
			return
		}

		// Make sure the hash exists with that specific file name!
		// This prevents people from faking wrong file names for downloading.
		fileHash := chi.URLParam(req, "hash")
		fileName := chi.URLParam(req, "file")
		fileSize := int64(0)
		found := false
		for _, file := range manifest.Files {
			if file.Object.Hash != fileHash {
				continue
			}
			if filepath.Base(file.Path) != fileName {
				continue
			}
			found = true
			fileSize = file.Object.Size
			break
		}

		if !found {
			http.Error(writer, "File not found", http.StatusNotFound)
			return
		}

		reader, err := packageStore.ReadObject(fileHash)
		if err != nil {
			log.Print(fmt.Errorf("error reading object %s from store: %w", fileHash, err))
			http.Error(writer, "Unable to read object", http.StatusInternalServerError)
			return
		}
		defer reader.Close()

		writer.Header().Set("Content-Type", "application/octet-stream")
		if fileSize > 1000 && strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
			writer.Header().Set("Content-Encoding", "gzip")
			gzipWriter := gzip.NewWriter(writer)
			defer gzipWriter.Close()
			_, err = io.Copy(gzipWriter, reader)
		} else {
			_, err = io.Copy(writer, reader)
		}

		if err != nil {
			log.Print(fmt.Errorf("error writing response data: %w", err))
			return
		}
	}
}
