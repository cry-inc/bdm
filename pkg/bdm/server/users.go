package server

type Roles struct {
	Reader bool
	Writer bool
}

type User struct {
	Id string
	Roles
}

type Users interface {
	ListUsers() ([]User, error)
	CreateUser(user User, password string) error
	DeleteUser(id string) error
	SetRoles(id string, roles *Roles) error
	GetRoles(id string) (*Roles, error)
	Authenticate(id, password string) bool
}
