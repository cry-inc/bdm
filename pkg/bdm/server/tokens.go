package server

import "time"

// Token describes a server token
type Token struct {
	Id         string
	Name       string
	Secret     string
	Expiration time.Time
	Roles
}

// The Tokens interface is used by the server as abstraction for token handling
type Tokens interface {
	CanRead(secret string) bool
	CanWrite(secret string) bool
	IsAdmin(secret string) bool

	GetTokens(userId string) ([]Token, error)
	CreateToken(userId, name string, expiration time.Time, roles *Roles) (*Token, error)
	DeleteToken(tokenId string) error
}
