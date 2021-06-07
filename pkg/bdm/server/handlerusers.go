package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func createUsersGetHandler(users Users) http.HandlerFunc {
	return enforceAdminUser(users, func(writer http.ResponseWriter, req *http.Request, authUser *User, paramUser *User) {
		userIds, err := users.GetUsers()
		if err != nil {
			log.Print(fmt.Errorf("error getting users list: %w", err))
			http.Error(writer, "Failed to get user list", http.StatusInternalServerError)
			return
		}

		userList := make([]*User, 0)
		for _, id := range userIds {
			user, err := users.GetUser(id)
			if err == nil {
				userList = append(userList, user)
			}
		}

		jsonData, err := json.Marshal(userList)
		if err != nil {
			log.Print(fmt.Errorf("error marshalling user list JSON: %w", err))
			http.Error(writer, "Failed to generate JSON user list", http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(jsonData)
	})
}

type createUserRequest struct {
	Id       string
	Password string
}

func createUsersPostHandler(users Users) http.HandlerFunc {
	return enforceSmallBodySize(enforceAdminUser(users, func(writer http.ResponseWriter, req *http.Request, authUser *User, paramUser *User) {
		jsonData, err := io.ReadAll(req.Body)
		if err != nil {
			log.Print(fmt.Errorf("error reading create user request: %w", err))
			http.Error(writer, "Failed read create user request", http.StatusBadRequest)
			return
		}

		var create createUserRequest
		err = json.Unmarshal(jsonData, &create)
		if err != nil {
			log.Print(fmt.Errorf("error unmarshalling JSON user data: %w", err))
			http.Error(writer, "Failed to parse JSON user data", http.StatusBadRequest)
			return
		}

		// Check for duplicate user ID
		_, err = users.GetUser(create.Id)
		if err == nil {
			http.Error(writer, "User ID is already existing", http.StatusConflict)
			return
		}

		newUser := User{Id: create.Id}
		err = users.CreateUser(newUser, create.Password)
		if err != nil {
			log.Print(fmt.Errorf("failed to create new user: %w", err))
			http.Error(writer, "Failed to create new user", http.StatusInternalServerError)
			return
		}

		jsonData, err = json.Marshal(newUser)
		if err != nil {
			log.Print(fmt.Errorf("error marshalling JSON user data: %w", err))
			http.Error(writer, "Failed to generate JSON user data", http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(jsonData)
	}))
}

func createUserGetHandler(users Users) http.HandlerFunc {
	return enforceAdminOrMatchUser(users, func(writer http.ResponseWriter, req *http.Request, authUser *User, paramUser *User) {
		jsonData, err := json.Marshal(paramUser)
		if err != nil {
			log.Print(fmt.Errorf("error marshalling JSON user data: %w", err))
			http.Error(writer, "Failed to generate JSON user data", http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(jsonData)
	})
}

func createUserDeleteHandler(users Users) http.HandlerFunc {
	return enforceAdminOrMatchUser(users, func(writer http.ResponseWriter, req *http.Request, authUser *User, paramUser *User) {
		err := users.DeleteUser(paramUser.Id)
		if err != nil {
			log.Print(fmt.Errorf("error deleting user: %w", err))
			http.Error(writer, "Failed to delete user", http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(writer, "null")
	})
}

type changePasswordRequest struct {
	OldPassword string
	NewPassword string
}

func createUserPatchPasswordHandler(users Users) http.HandlerFunc {
	return enforceSmallBodySize(enforceAdminOrMatchUser(users, func(writer http.ResponseWriter, req *http.Request, authUser *User, paramUser *User) {
		jsonData, err := io.ReadAll(req.Body)
		if err != nil {
			log.Print(fmt.Errorf("error reading user patch request: %w", err))
			http.Error(writer, "Failed read user change request", http.StatusBadRequest)
			return
		}

		var passChange changePasswordRequest
		err = json.Unmarshal(jsonData, &passChange)
		if err != nil {
			http.Error(writer, "Failed to parse JSON password data", http.StatusBadRequest)
			return
		}

		// Admins can change passwords for others, otherwise the old PW must be provided
		if !authUser.Admin || authUser.Id == paramUser.Id {
			if !users.Authenticate(paramUser.Id, passChange.OldPassword) {
				http.Error(writer, "Old password does not match", http.StatusBadRequest)
				return
			}
		}

		err = users.ChangePassword(paramUser.Id, passChange.NewPassword)
		if err != nil {
			http.Error(writer, "Failed to apply new password", http.StatusBadRequest)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write([]byte("{}"))
	}))
}

type changeRolesRequest struct {
	Roles
}

func createUserPatchRolesHandler(users Users) http.HandlerFunc {
	return enforceSmallBodySize(enforceAdminUser(users, func(writer http.ResponseWriter, req *http.Request, authUser *User, paramUser *User) {
		jsonData, err := io.ReadAll(req.Body)
		if err != nil {
			log.Print(fmt.Errorf("error reading user patch request: %w", err))
			http.Error(writer, "Failed read user change request", http.StatusBadRequest)
			return
		}

		var roleChange changeRolesRequest
		err = json.Unmarshal(jsonData, &roleChange)
		if err != nil {
			http.Error(writer, "Failed to parse JSON role data", http.StatusBadRequest)
			return
		}

		err = users.SetRoles(paramUser.Id, &roleChange.Roles)
		if err != nil {
			log.Print(fmt.Errorf("failed to set new roles: %w", err))
			http.Error(writer, "Failed to apply new roles", http.StatusInternalServerError)
			return
		}

		changedUser, err := users.GetUser(paramUser.Id)
		if err != nil {
			log.Print(fmt.Errorf("changed user no longer exists: %w", err))
			http.Error(writer, "Changed user no longer exists", http.StatusInternalServerError)
			return
		}

		jsonData, err = json.Marshal(changedUser.Roles)
		if err != nil {
			log.Print(fmt.Errorf("error marshalling JSON role data: %w", err))
			http.Error(writer, "Failed to generate JSON role data", http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(jsonData)
	}))
}
