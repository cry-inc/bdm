package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type censoredToken struct {
	Id         string
	Name       string
	Expiration time.Time
	Roles
}

func createTokensGetHandler(users Users, tokens Tokens) http.HandlerFunc {
	return enforceAdminOrMatchUser(users, func(writer http.ResponseWriter, req *http.Request, authUser *User, paramUser *User) {
		tokenList, err := tokens.GetTokens(paramUser.Id)
		if err != nil {
			log.Print(fmt.Errorf("error getting token list: %w", err))
			http.Error(writer, "Failed to get token list", http.StatusInternalServerError)
			return
		}

		censoredList := make([]censoredToken, 0)
		for _, t := range tokenList {
			censoredToken := censoredToken{
				Id:         t.Id,
				Name:       t.Name,
				Expiration: t.Expiration,
				Roles:      t.Roles,
			}
			censoredList = append(censoredList, censoredToken)
		}

		jsonData, err := json.Marshal(censoredList)
		if err != nil {
			log.Print(fmt.Errorf("error marshalling token list to JSON: %w", err))
			http.Error(writer, "Failed to generate JSON for token list", http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(jsonData)
	})
}

type createTokenRequest struct {
	Name       string
	Expiration time.Time
	Roles
}

func createTokensPostHandler(users Users, tokens Tokens) http.HandlerFunc {
	return enforceSmallBodySize(enforceAdminOrMatchUser(users, func(writer http.ResponseWriter, req *http.Request, authUser *User, paramUser *User) {
		tokenList, err := tokens.GetTokens(paramUser.Id)
		if err != nil {
			log.Print(fmt.Errorf("error reading existing user tokens: %w", err))
			http.Error(writer, "Failed to list existing user tokens", http.StatusInternalServerError)
			return
		}

		const maxUserTokens = 10
		if len(tokenList) >= maxUserTokens {
			http.Error(writer, "Exceeded limit of tokens per user", http.StatusBadRequest)
			return
		}

		jsonData, err := io.ReadAll(req.Body)
		if err != nil {
			log.Print(fmt.Errorf("error reading create token request: %w", err))
			http.Error(writer, "Failed to read create token request", http.StatusBadRequest)
			return
		}

		createRequest := createTokenRequest{}
		err = json.Unmarshal(jsonData, &createRequest)
		if err != nil {
			log.Print(fmt.Errorf("error unmarshalling JSON role data: %w", err))
			http.Error(writer, "Failed to parse JSON role data", http.StatusBadRequest)
			return
		}

		// Make sure the target user has the requested permissions
		invalidAdminRequest := createRequest.Admin && !paramUser.Admin
		invalidWriterRequest := createRequest.Writer && !paramUser.Writer
		invalidReaderRequest := createRequest.Reader && !paramUser.Reader
		if invalidAdminRequest || invalidWriterRequest || invalidReaderRequest {
			http.Error(writer, "Requested invalid role", http.StatusForbidden)
			return
		}

		// Make sure there is at least one role
		if !createRequest.Admin && !createRequest.Writer && !createRequest.Reader {
			http.Error(writer, "Requested no role", http.StatusBadRequest)
			return
		}

		// Make sure there is a name with length between 1 and 255
		if len(createRequest.Name) < 1 || len(createRequest.Name) > 255 {
			http.Error(writer, "Invalid token name", http.StatusBadRequest)
			return
		}

		// Make sure the expiration date is in the future
		if createRequest.Expiration.Before(time.Now()) {
			http.Error(writer, "Invalid token expiration", http.StatusBadRequest)
			return
		}

		token, err := tokens.CreateToken(paramUser.Id, createRequest.Name, createRequest.Expiration, &createRequest.Roles)
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
