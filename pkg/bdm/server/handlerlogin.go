package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type loginRequest struct {
	UserId   string
	Password string
}

func createLoginGetHandler(users Users) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		if !users.Available() {
			http.Error(writer, "User system is disabled", http.StatusServiceUnavailable)
			return
		}

		cookie, err := req.Cookie("login")
		if err != nil {
			writer.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(writer, "null")
			return
		}

		token, err := readAuthToken(cookie.Value)
		if err != nil {
			writer.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(writer, "null")
			return
		}

		user, err := users.GetUser(token.UserId)
		if err != nil {
			http.Error(writer, "Failed to find user", http.StatusNotFound)
			return
		}

		jsonData, err := json.Marshal(user)
		if err != nil {
			log.Print(fmt.Errorf("error marshalling login JSON response: %w", err))
			http.Error(writer, "Failed to generate JSON", http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(jsonData)
	}
}

func createLoginPostHandler(users Users) http.HandlerFunc {
	return enforceSmallBodySize(func(writer http.ResponseWriter, req *http.Request) {
		if !users.Available() {
			http.Error(writer, "User system is disabled", http.StatusServiceUnavailable)
			return
		}

		jsonData, err := io.ReadAll(req.Body)
		if err != nil {
			log.Print(fmt.Errorf("error reading login request: %w", err))
			http.Error(writer, "Failed to read login request", http.StatusInternalServerError)
			return
		}

		var login loginRequest
		err = json.Unmarshal(jsonData, &login)
		if err != nil {
			log.Print(fmt.Errorf("error unmarshalling JSON login data: %w", err))
			http.Error(writer, "Failed to parse JSON", http.StatusInternalServerError)
			return
		}

		valid := users.Authenticate(login.UserId, login.Password)
		if !valid {
			http.Error(writer, "Failed to log in", http.StatusUnauthorized)
			return
		}

		authToken := createAuthToken(login.UserId, defaultExpiration)
		cookie := http.Cookie{
			Name:     "login",
			Value:    authToken.Token,
			Expires:  authToken.Expires,
			SameSite: http.SameSiteStrictMode,
			HttpOnly: true,
		}
		http.SetCookie(writer, &cookie)

		user, err := users.GetUser(login.UserId)
		if err != nil {
			http.Error(writer, "Failed to find user", http.StatusNotFound)
			return
		}

		jsonData, err = json.Marshal(user)
		if err != nil {
			log.Print(fmt.Errorf("error marshalling login JSON response: %w", err))
			http.Error(writer, "Failed to generate JSON", http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(jsonData)
	})
}

func createLoginDeleteHandler(users Users) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		if !users.Available() {
			http.Error(writer, "User system is disabled", http.StatusServiceUnavailable)
			return
		}

		cookie := http.Cookie{
			Name:     "login",
			Value:    "",
			SameSite: http.SameSiteStrictMode,
			HttpOnly: true,
		}
		http.SetCookie(writer, &cookie)
		writer.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(writer, "null")
	}
}
