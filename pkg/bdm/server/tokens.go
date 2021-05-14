package server

type Token struct {
	Id string
	Roles
}

type Tokens interface {
	CanRead(token string) bool
	CanWrite(token string) bool
	IsAdmin(token string) bool

	NoUserMode() bool
	GetTokens(userId string) ([]Token, error)
	CreateToken(userId string, roles *Roles) (*Token, error)
	DeleteToken(tokenId string) error
}
