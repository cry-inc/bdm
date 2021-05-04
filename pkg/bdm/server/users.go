package server

type Roles struct {
	Reader bool
	Writer bool
}

type User struct {
	Id string
	Roles
}

type Token struct {
	Id string
	Roles
}

type Users interface {
	GetUsers() ([]User, error)
	CreateUser(user User, password string) error
	DeleteUser(userId string) error
	SetRoles(userId string, roles *Roles) error
	GetRoles(userId string) (*Roles, error)
	Authenticate(userId, password string) bool
	ChangePassword(userId, password string) error
	GetTokens(userId string) ([]Token, error)
	CreateToken(userId string, roles *Roles) (*Token, error)
	DeleteToken(tokenId string) error
	Permissions
}
