package server

import (
	"fmt"
	"net/http"

	"github.com/cry-inc/bdm/pkg/bdm"
	"github.com/go-chi/chi/v5"
)

// Ensures that clients do not send huge bodies to create server issues
func enforceMaxBodySize(handler http.HandlerFunc, maxSize int64) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		req.Body = http.MaxBytesReader(writer, req.Body, maxSize)
		handler(writer, req)
	}
}

// Limit request bodies for requests with small JSON payloads
func enforceSmallBodySize(handler http.HandlerFunc) http.HandlerFunc {
	const maxSmallBodySize = 1024 * 100 // 100 kB is enough for small JSON payloads
	return enforceMaxBodySize(handler, maxSmallBodySize)
}

// Limit request bodies for requests with bigger JSON payloads
func enforceJsonBodySize(handler http.HandlerFunc) http.HandlerFunc {
	// Use size limit from base package that is used everywhere for JSON data
	return enforceMaxBodySize(handler, bdm.JsonSizeLimit)
}

// Checks if a request contains permissions for reading.
// Checks for an BDM API token (typically used by the CLI client) with read permissions
// or for auth tokens (used by the Web UI logins) of users with read permissions.
func hasReadPermission(request *http.Request, users Users, tokens Tokens) bool {
	apiToken := request.Header.Get(bdm.ApiTokenHeader)
	if tokens.CanRead(apiToken) {
		return true
	}
	user, err := getCurrentUser(request, users)
	return err == nil && user.Reader
}

// Checks if a request contains permissions for writing.
// Checks for an BDM API token (typically used by the CLI client) with write permissions
// or for auth tokens (used by the Web UI logins) of users with write permissions.
func hasWritePermission(request *http.Request, users Users, tokens Tokens) bool {
	apiToken := request.Header.Get(bdm.ApiTokenHeader)
	if tokens.CanWrite(apiToken) {
		return true
	}
	user, err := getCurrentUser(request, users)
	return err == nil && user.Writer
}

// Extracts the logged in Web UI user identified by an auth token from an incoming request
func getCurrentUser(request *http.Request, users Users) (*User, error) {
	cookie, err := request.Cookie("login")
	if err != nil {
		return nil, fmt.Errorf("no login cookie found: %w", err)
	}
	token, err := readAuthToken(cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("unable to read token: %w", err)
	}
	user, err := users.GetUser(token.UserId)
	if err != nil {
		return nil, fmt.Errorf("unable to find user: %w", err)
	}
	return user, nil
}

type UserHandlerFunc func(writer http.ResponseWriter, req *http.Request, authUser *User, paramUser *User)

// Wrapper for http.HandlerFunc that enforces and looks up a logged in user.
// Additionally it also extracts the user ID from the URL paramater.
// The authenticated users is required or leads to an error.
// The URL parameter user is optional and can be nil for routes without the parameter.
// Both users are handed over as *User arguments.
func extractUsers(users Users, handler UserHandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		authUser, err := getCurrentUser(req, users)
		if err != nil {
			http.Error(writer, "Log in required", http.StatusForbidden)
			return
		}

		var paramUser *User = nil
		paramUserId := chi.URLParam(req, "user")
		if len(paramUserId) > 0 {
			paramUser, err = users.GetUser(paramUserId)
			if err != nil {
				http.Error(writer, "User from URL does not exist", http.StatusNotFound)
				return
			}
		}

		handler(writer, req, authUser, paramUser)
	}
}

// Wrapper around extractUsers that enforces that the authenticated user is an admin
func enforceAdminUser(users Users, handler UserHandlerFunc) http.HandlerFunc {
	return extractUsers(users, func(writer http.ResponseWriter, req *http.Request, authUser *User, paramUser *User) {
		if !authUser.Admin {
			http.Error(writer, "Admin permissions required", http.StatusUnauthorized)
			return
		}

		handler(writer, req, authUser, paramUser)
	})
}

// Wrapper around extractUsers that enforces that the authenticated user is an admin OR
// the autheticated user matches the user from the URL parameters.
func enforceAdminOrMatchUser(users Users, handler UserHandlerFunc) http.HandlerFunc {
	return extractUsers(users, func(writer http.ResponseWriter, req *http.Request, authUser *User, paramUser *User) {
		if paramUser != nil && authUser.Id != paramUser.Id && !authUser.Admin {
			http.Error(writer, "Admin permissions required", http.StatusUnauthorized)
			return
		}

		handler(writer, req, authUser, paramUser)
	})
}
