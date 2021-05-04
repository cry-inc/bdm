package server

type Roles struct {
	Reader bool
	Writer bool
}

type User struct {
	Id string
	Roles
}

const ReadToken = "Read"
const WriteToken = "Write"

type Token struct {
	Id   string
	Type string
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
	CreateToken(userId, tokenType string) (string, error)
	DeleteToken(tokenId string) error
}
