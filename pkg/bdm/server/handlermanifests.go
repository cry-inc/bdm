package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/cry-inc/bdm/pkg/bdm"
	"github.com/cry-inc/bdm/pkg/bdm/store"
	"github.com/julienschmidt/httprouter"
)

func createManifestHandler(packageStore store.Store) httprouter.Handle {
	return func(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
		name := params.ByName("name")
		validName := bdm.ValidatePackageName(name)
		if !validName {
			http.Error(writer, "Bad package name", http.StatusBadRequest)
			return
		}

		versionString := params.ByName("version")
		version, err := strconv.Atoi(versionString)
		if err != nil || version <= 0 {
			http.Error(writer, "Bad package version", http.StatusBadRequest)
			return
		}

		manifest, err := packageStore.GetManifest(name, uint(version))
		if err != nil {
			http.Error(writer, "Package does not exist", http.StatusNotFound)
			return
		}

		jsonData, err := json.Marshal(*manifest)
		if err != nil {
			log.Print(fmt.Errorf("error marshalling manifest to JSON: %w", err))
			http.Error(writer, "Failed to generate JSON data", http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(jsonData)
	}
}

func createManifestNamesHandler(packageStore store.Store) httprouter.Handle {
	return func(writer http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		names, err := packageStore.GetNames()
		if err != nil {
			log.Print(fmt.Errorf("error listing package names: %w", err))
			http.Error(writer, "Failed to list package names", http.StatusInternalServerError)
			return
		}

		type manifestListItem struct{ Name string }
		manifestList := make([]manifestListItem, 0)
		for _, name := range names {
			manifestList = append(manifestList, manifestListItem{Name: name})
		}

		jsonData, err := json.Marshal(manifestList)
		if err != nil {
			log.Print(fmt.Errorf("error marshalling manifest to JSON: %w", err))
			http.Error(writer, "Failed to generate JSON data", http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(jsonData)
	}
}

func createManifestVersionsHandler(packageStore store.Store) httprouter.Handle {
	return func(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
		name := params.ByName("name")
		validName := bdm.ValidatePackageName(name)
		if !validName {
			http.Error(writer, "Bad package name", http.StatusBadRequest)
			return
		}

		versions, err := packageStore.GetVersions(name)
		if err != nil {
			log.Print(fmt.Errorf("error getting version numbers for package %s: %w", name, err))
			http.Error(writer, "Failed to list package versions", http.StatusInternalServerError)
			return
		}
		if len(versions) == 0 {
			http.Error(writer, "Package not found", http.StatusNotFound)
			return
		}

		type versionListItem struct{ Version uint }
		versionList := make([]versionListItem, 0)
		for _, version := range versions {
			versionList = append(versionList, versionListItem{Version: version})
		}

		jsonData, err := json.Marshal(versionList)
		if err != nil {
			log.Print(fmt.Errorf("error marshalling version numbers to JSON: %w", err))
			http.Error(writer, "Failed to generate JSON data", http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(jsonData)
	}
}

func createPublishManifestHandler(packageStore store.Store, limits *bdm.ManifestLimits, apiKey string) httprouter.Handle {
	return func(writer http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		requestAPIKey := req.Header.Get(apiKeyField)
		if len(requestAPIKey) == 0 || apiKey != requestAPIKey {
			http.Error(writer, "Wrong API key", http.StatusUnauthorized)
			return
		}

		jsonData, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(writer, "Bad request", http.StatusBadRequest)
			return
		}

		var manifest bdm.Manifest
		err = json.Unmarshal(jsonData, &manifest)
		if err != nil {
			http.Error(writer, "Bad JSON data", http.StatusBadRequest)
			return
		}

		err = bdm.ValidateUnpublishedManifest(&manifest)
		if err != nil {
			http.Error(writer, "Bad manifest", http.StatusBadRequest)
			return
		}

		err = bdm.CheckManifestLimits(&manifest, limits)
		if err != nil {
			http.Error(writer, "Manifest exceeds server limits", http.StatusBadRequest)
			return
		}

		allObjectsExist := store.AllObjectsExist(&manifest, packageStore)
		if !allObjectsExist {
			http.Error(writer, "Not all manifest objects exist", http.StatusBadRequest)
			return
		}

		err = packageStore.PublishManifest(&manifest)
		var dupErr store.DuplicatePackageError
		if errors.As(err, &dupErr) {
			http.Error(writer, "Older package with same content exists already", http.StatusConflict)
			return
		}
		if err != nil {
			http.Error(writer, "Failed to publish manifest", http.StatusInternalServerError)
			return
		}

		jsonData, err = json.Marshal(manifest)
		if err != nil {
			http.Error(writer, "Failed to serialize published manifest", http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(jsonData)
	}
}
