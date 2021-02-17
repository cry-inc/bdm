package server

import (
	"net/http"

	"github.com/cry-inc/bdm/pkg/bdm/store"
	"github.com/julienschmidt/httprouter"
)

const apiKeyField = "bdm-api-key"

// CreateRouter creates a new HTTP handler that handles all server routes
func CreateRouter(packageStore store.Store, apiKey string) (http.Handler, error) {
	router := httprouter.New()

	// Static assets for HTML UI
	staticHandler := createStaticHandler()
	router.GET("/", staticHandler)
	router.GET("/favicon.ico", staticHandler)

	// Download package files as ZIP
	router.GET("/zip/:name/:version", createZipHandler(packageStore))

	// Publish manifest for package
	router.POST("/manifests", createPublishManifestHandler(packageStore, apiKey))

	// Get list of package names
	router.GET("/manifests", createManifestNamesHandler(packageStore))

	// Get versions for specific package by name
	router.GET("/manifests/:name", createManifestVersionsHandler(packageStore))

	// Get manifest for specific package & version
	router.GET("/manifests/:name/:version", createManifestHandler(packageStore))

	// Upload one or more objects. The compressed request body contains:
	// - 8 bytes uint for JSON data length
	// - JSON data with bdm.Object array
	// - object data
	// The response body contains the uploaded objects as JSON array.
	router.POST("/objects/upload", createUploadObjectsHandler(packageStore, apiKey))

	// Check for existing objects. The request body contains:
	// - 8 bytes uint for JSON data length
	// - compressed JSON data with bdm.Object array
	// The response body contains the found objects as JSON array.
	router.POST("/objects/check", createCheckObjectsHandler(packageStore))

	// Download objects. The request body contains:
	// - 8 bytes uint for JSON data length
	// - compressed JSON data with bdm.Object array
	// The response body contains:
	// - 8 bytes uint for JSON data length
	// - compressed JSON data with bdm.Object array
	// - compressed object data
	router.POST("/objects/download", createDownloadObjectsHandler(packageStore))

	// Downloads a single file from a package
	router.GET("/files/:name/:version/:hash/:file", createFilesHandler(packageStore))

	return router, nil
}
