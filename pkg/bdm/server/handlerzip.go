package server

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/cry-inc/bdm/pkg/bdm"
	"github.com/cry-inc/bdm/pkg/bdm/store"
	"github.com/go-chi/chi/v5"
)

func createZipHandler(store store.Store, tokens Tokens) http.HandlerFunc {
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

		writer.Header().Set("Content-Type", "application/zip")
		writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.v%d.zip\"", name, version))

		err = streamPackageZip(name, uint(version), store, writer)
		if err != nil {
			log.Print(fmt.Errorf("error streaming zip package: %w", err))
			http.Error(writer, "Failed to stream ZIP data", http.StatusInternalServerError)
			return
		}
	}
}

func streamPackageZip(name string, version uint, store store.Store, output io.Writer) error {
	manifest, err := store.GetManifest(name, version)
	if err != nil {
		return fmt.Errorf("error getting manifest %s in version %d: %w", name, version, err)
	}

	zipWriter := zip.NewWriter(output)
	defer zipWriter.Close()

	for _, file := range manifest.Files {
		objectReader, err := store.ReadObject(file.Object.Hash)
		if err != nil {
			return fmt.Errorf("error reading object %s: %w", file.Object.Hash, err)
		}
		defer objectReader.Close()

		zipFile, err := zipWriter.Create(file.Path)
		if err != nil {
			return fmt.Errorf("error creating file %s in ZIP: %w", file.Path, err)
		}

		copyCount, err := io.Copy(zipFile, objectReader)
		if err != nil {
			return fmt.Errorf("error copying file %s into ZIP: %w", file.Path, err)
		}
		if copyCount != file.Object.Size {
			return fmt.Errorf("error copying file %s into ZIP: copied %d of %d bytes",
				file.Path, copyCount, file.Object.Size)
		}
	}

	return nil
}
