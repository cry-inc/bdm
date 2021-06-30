package server

import "time"

type Token struct {
	Id         string
	Name       string
	Secret     string
	Expiration time.Time
	Roles
}

type Tokens interface {
	CanRead(secret string) bool
	CanWrite(secret string) bool
	IsAdmin(secret string) bool

	GetTokens(userId string) ([]Token, error)
	CreateToken(userId, name string, expiration time.Time, roles *Roles) (*Token, error)
	DeleteToken(tokenId string) error
}
