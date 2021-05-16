package server

import (
	"fmt"
	"net/http"
)

type User struct {
	Id string
	Roles
}

// The Users interface represents a user database that bundles all user management
// functionailty for the servers. Since the server can be also used without individual
// users accounts, its possible Available() returns false. This means users are
// disabled and all other methods will always return an error.
type Users interface {
	Available() bool

	GetUsers() ([]string, error)

	CreateUser(user User, password string) error
	Authenticate(userId, password string) bool
	ChangePassword(userId, password string) error
	GetUser(userId string) (*User, error)
	DeleteUser(userId string) error

	SetRoles(userId string, roles *Roles) error
	GetRoles(userId string) (*Roles, error)
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
