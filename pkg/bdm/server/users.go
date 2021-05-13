package server

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
