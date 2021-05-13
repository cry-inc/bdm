package server

import (
	"fmt"
)

type noUsers struct{}

func CreateNoUsers() Users {
	return &noUsers{}
}

func (users *noUsers) Available() bool {
	return false
}

func (users *noUsers) GetUsers() ([]string, error) {
	return nil, fmt.Errorf("does not support users")
}

func (users *noUsers) GetUser(userId string) (*User, error) {
	return nil, fmt.Errorf("does not support users")
}

func (users *noUsers) CreateUser(user User, password string) error {
	return fmt.Errorf("does not support users")
}

func (users *noUsers) DeleteUser(userId string) error {
	return fmt.Errorf("does not support users")
}

func (users *noUsers) GetRoles(userId string) (*Roles, error) {
	return nil, fmt.Errorf("does not support users")
}

func (users *noUsers) SetRoles(userId string, roles *Roles) error {
	return fmt.Errorf("does not support users")
}

func (users *noUsers) Authenticate(userId, password string) bool {
	return false
}

func (users *noUsers) ChangePassword(userId, password string) error {
	return fmt.Errorf("does not support users")
}
