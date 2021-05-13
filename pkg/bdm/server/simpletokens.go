package server

import "fmt"

type simpleTokens struct {
	readToken  string
	writeToken string
}

// SimpleTokens returns a simple token implementation that allows reading
// and uploading based on two single shared secret tokens. An empty token means no
// permission required and everyone is allowed for the corresponding action.
// Please keep in mind that a writing token will also always grant read permission!
func SimpleTokens(readToken, writeToken string) Tokens {
	tokens := simpleTokens{readToken, writeToken}
	return &tokens
}

func (s *simpleTokens) CanRead(token string) bool {
	return token == s.readToken || token == s.writeToken
}

func (s *simpleTokens) CanWrite(token string) bool {
	return token == s.writeToken
}

func (s *simpleTokens) NoUserMode() bool {
	return true
}

func (s *simpleTokens) GetTokens(userId string) ([]Token, error) {
	return nil, fmt.Errorf("does not support users")
}

func (s *simpleTokens) CreateToken(userId string, roles *Roles) (*Token, error) {
	return nil, fmt.Errorf("does not support users")
}

func (s *simpleTokens) DeleteToken(tokenId string) error {
	return fmt.Errorf("does not support deletion")
}
