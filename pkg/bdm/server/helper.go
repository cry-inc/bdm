package server

import (
	"fmt"
	"net/http"
)

const apiTokenField = "bdm-api-token"

func hasReadToken(request *http.Request, tokens Tokens) bool {
	apiToken := request.Header.Get(apiTokenField)
	return tokens.CanRead(apiToken)
}

func hasWriteToken(request *http.Request, tokens Tokens) bool {
	apiToken := request.Header.Get(apiTokenField)
	return tokens.CanWrite(apiToken)
}

func getCurrentUser(request *http.Request, users Users) (*User, error) {
	if !users.Available() {
		return nil, fmt.Errorf("user support is disabled")
	}
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
