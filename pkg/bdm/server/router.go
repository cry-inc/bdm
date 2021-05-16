package server

import (
	"net/http"

	"github.com/cry-inc/bdm/pkg/bdm"
	"github.com/cry-inc/bdm/pkg/bdm/store"
	"github.com/go-chi/chi/v5"
)

const apiTokenField = "bdm-api-token"

// CreateRouter creates a new HTTP handler that handles all server routes
func CreateRouter(packageStore store.Store, limits *bdm.ManifestLimits, users Users, tokens Tokens) http.Handler {
	router := chi.NewRouter()

	// Static assets for HTML UI
	staticHandler := createStaticHandler()
	router.Get("/", staticHandler)
	router.Get("/favicon.ico", staticHandler)

	// Download package files as ZIP
	router.Get("/zip/{name}/{version}", createZipHandler(packageStore, tokens))

	// Publish manifest for package
	router.Get("/limits", createLimitsHandler(limits, tokens))

	// Publish manifest for package
	router.Post("/manifests", createPublishManifestHandler(packageStore, limits, tokens))

	// Get list of package names
	router.Get("/manifests", createManifestNamesHandler(packageStore, tokens))

	// Get versions for specific package by name
	router.Get("/manifests/{name}", createManifestVersionsHandler(packageStore, tokens))

	// Get manifest for specific package & version
	router.Get("/manifests/{name}/{version}", createManifestHandler(packageStore, tokens))

	// Upload one or more objects. The compressed request body contains:
	// - 8 bytes uint for JSON data length
	// - JSON data with bdm.Object array
	// - object data
	// The response body contains the uploaded objects as JSON array.
	router.Post("/objects/upload", createUploadObjectsHandler(packageStore, limits, tokens))

	// Check for existing objects. The request body contains:
	// - 8 bytes uint for JSON data length
	// - compressed JSON data with bdm.Object array
	// The response body contains the found objects as JSON array.
	router.Post("/objects/check", createCheckObjectsHandler(packageStore, tokens))

	// Download objects. The request body contains:
	// - 8 bytes uint for JSON data length
	// - compressed JSON data with bdm.Object array
	// The response body contains:
	// - 8 bytes uint for JSON data length
	// - compressed JSON data with bdm.Object array
	// - compressed object data
	router.Post("/objects/download", createDownloadObjectsHandler(packageStore, tokens))

	// Downloads a single file from a package
	router.Get("/files/{name}/{version}/{hash}/{file}", createFilesHandler(packageStore, tokens))

	// Login
	router.Post("/login", createLoginPostHandler(users))
	// Logout
	router.Delete("/login", createLoginDeleteHandler(users))
	// Get current user
	router.Get("/login", createLoginGetHandler(users))

	// List all users
	router.Get("/users", createUsersGetHandler(users))
	// Create new user
	router.Post("/users", createUsersPostHandler(users))
	// Get specific user
	router.Get("/users/{user}", createUserGetHandler(users))
	// Delete specific user
	router.Delete("/users/{user}", createUserDeleteHandler(users))
	// Change user PW
	router.Patch("/users/{user}/password", createUserPatchPasswordHandler(users))
	// Change user roles
	router.Patch("/users/{user}/roles", createUserPatchRolesHandler(users))

	return router
}
