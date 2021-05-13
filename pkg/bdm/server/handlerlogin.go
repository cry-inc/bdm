package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type loginRequest struct {
	UserId   string
	Password string
}

type loginResponse struct {
	UserId string
}

func createLoginGetHandler(users Users) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		if !users.Available() {
			http.Error(writer, "User system is disabled", http.StatusForbidden)
			return
		}

		cookie, err := req.Cookie("login")
		if err != nil {
			writer.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(writer, `{"UserId": null}`)
			return
		}

		token, err := readAuthToken(cookie.Value)
		if err != nil {
			writer.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(writer, `{"UserId": null}`)
			return
		}

		response := loginResponse{token.UserId}
		jsonData, err := json.Marshal(response)
		if err != nil {
			log.Print(fmt.Errorf("error marshalling login JSON response: %w", err))
			http.Error(writer, "Failed to generate JSON login data", http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(jsonData)
	}
}

func createLoginPostHandler(users Users) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		if !users.Available() {
			http.Error(writer, "User system is disabled", http.StatusForbidden)
			return
		}

		jsonData, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Print(fmt.Errorf("error reading login request: %w", err))
			http.Error(writer, "Failed read login request", http.StatusInternalServerError)
			return
		}

		var login loginRequest
		err = json.Unmarshal(jsonData, &login)
		if err != nil {
			log.Print(fmt.Errorf("error unmarshalling JSON login data: %w", err))
			http.Error(writer, "Failed to parse JSON login data", http.StatusInternalServerError)
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

		response := loginResponse{login.UserId}
		jsonData, err = json.Marshal(response)
		if err != nil {
			log.Print(fmt.Errorf("error marshalling login JSON response: %w", err))
			http.Error(writer, "Failed to generate JSON login data", http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(jsonData)
	}
}

func createLoginDeleteHandler(users Users) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		if !users.Available() {
			http.Error(writer, "User system is disabled", http.StatusForbidden)
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
		fmt.Fprintf(writer, `{"UserId": null}`)
	}
}
