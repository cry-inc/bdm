package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func createTokensGetHandler(users Users, tokens Tokens) http.HandlerFunc {
	return enforceAdminOrMatchUser(users, func(writer http.ResponseWriter, req *http.Request, authUser *User, paramUser *User) {
		tokenList, err := tokens.GetTokens(paramUser.Id)
		if err != nil {
			log.Print(fmt.Errorf("error getting token list: %w", err))
			http.Error(writer, "Failed to get token list", http.StatusInternalServerError)
			return
		}

		jsonData, err := json.Marshal(tokenList)
		if err != nil {
			log.Print(fmt.Errorf("error marshalling token list to JSON: %w", err))
			http.Error(writer, "Failed to generate JSON for token list", http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(jsonData)
	})
}

func createTokensPostHandler(users Users, tokens Tokens) http.HandlerFunc {
	return enforceSmallBodySize(enforceAdminOrMatchUser(users, func(writer http.ResponseWriter, req *http.Request, authUser *User, paramUser *User) {
		jsonData, err := io.ReadAll(req.Body)
		if err != nil {
			log.Print(fmt.Errorf("error reading create token request: %w", err))
			http.Error(writer, "Failed read create token request", http.StatusBadRequest)
			return
		}

		tokenRoles := Roles{}
		err = json.Unmarshal(jsonData, &tokenRoles)
		if err != nil {
			log.Print(fmt.Errorf("error unmarshalling JSON role data: %w", err))
			http.Error(writer, "Failed to parse JSON role data", http.StatusBadRequest)
			return
		}

		// Make sure the target user has the requested permissions
		invalidAdminRequest := tokenRoles.Admin && !paramUser.Admin
		invalidWriterRequest := tokenRoles.Writer && !paramUser.Writer
		invalidReaderRequest := tokenRoles.Reader && !paramUser.Reader
		if invalidAdminRequest || invalidWriterRequest || invalidReaderRequest {
			http.Error(writer, "Requested invalid role", http.StatusForbidden)
			return
		}

		// Make sure there is at least one role
		if !tokenRoles.Admin && !tokenRoles.Writer && !tokenRoles.Reader {
			http.Error(writer, "Requested no role", http.StatusBadRequest)
			return
		}

		token, err := tokens.CreateToken(paramUser.Id, &tokenRoles)
		if err != nil {
			log.Print(fmt.Errorf("failed to create new token: %w", err))
			http.Error(writer, "Failed to create new token", http.StatusInternalServerError)
			return
		}

		jsonData, err = json.Marshal(token)
		if err != nil {
			log.Print(fmt.Errorf("error marshalling token to JSON: %w", err))
			http.Error(writer, "Failed to generate JSON for token", http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(jsonData)
	}))
}

func createTokensDeleteHandler(users Users, tokens Tokens) http.HandlerFunc {
	return enforceAdminOrMatchUser(users, func(writer http.ResponseWriter, req *http.Request, authUser *User, paramUser *User) {
		tokenId := chi.URLParam(req, "token")

		tokenList, err := tokens.GetTokens(paramUser.Id)
		if err != nil {
			log.Print(fmt.Errorf("error getting token list: %w", err))
			http.Error(writer, "Failed to get token list", http.StatusInternalServerError)
			return
		}

		belongsToParamUser := false
		for _, t := range tokenList {
			if t.Id == tokenId {
				belongsToParamUser = true
				break
			}
		}

		if !belongsToParamUser {
			http.Error(writer, "Requested invalid token deletion", http.StatusForbidden)
			return
		}

		err = tokens.DeleteToken(tokenId)
		if err != nil {
			log.Print(fmt.Errorf("error deleting token: %w", err))
			http.Error(writer, "Failed to delete token", http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(writer, "{}")
	})
}
