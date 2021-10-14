package server

// User describes a server user
type User struct {
	Id string
	Roles
}

// The Users interface is used by the server as abstraction for user management
type Users interface {
	GetUsers() ([]string, error)

	CreateUser(user User, password string) error
	Authenticate(userId, password string) bool
	ChangePassword(userId, password string) error
	GetUser(userId string) (*User, error)
	DeleteUser(userId string) error

	SetRoles(userId string, roles *Roles) error
	GetRoles(userId string) (*Roles, error)
}
