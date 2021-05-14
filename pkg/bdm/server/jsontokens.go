package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"github.com/cry-inc/bdm/pkg/bdm/util"
)

type jsonToken struct {
	Token
	UserId string
}

type jsonTokens struct {
	tokensFile string
	tokens     map[string]jsonToken
	mutex      sync.Mutex
	users      Users
}

func CreateJsonTokens(tokensFile string, users Users) (Tokens, error) {
	if !users.Available() {
		return nil, fmt.Errorf("user management does not support individual users")
	}

	tokens := jsonTokens{
		tokensFile: tokensFile,
		tokens:     make(map[string]jsonToken),
		users:      users,
	}

	if !util.FileExists(tokens.tokensFile) {
		err := tokens.saveTokens()
		if err != nil {
			return nil, fmt.Errorf("unable to create token database file %s: %w", tokens.tokensFile, err)
		}
	}

	err := tokens.loadTokens()
	if err != nil {
		return nil, fmt.Errorf("unable to load token database: %w", err)
	}

	return &tokens, nil
}

func (tokens *jsonTokens) loadTokens() error {
	jsonData, err := ioutil.ReadFile(tokens.tokensFile)
	if err != nil {
		return fmt.Errorf("error reading token database file %s: %w", tokens.tokensFile, err)
	}

	var tokenList []jsonToken
	err = json.Unmarshal(jsonData, &tokenList)
	if err != nil {
		return fmt.Errorf("error while unmarshalling token database: %w", err)
	}

	tokens.tokens = make(map[string]jsonToken)
	for _, t := range tokenList {
		tokens.tokens[t.Id] = t
	}

	return nil
}

func (tokens *jsonTokens) saveTokens() error {
	tokenList := make([]jsonToken, 0)
	for _, t := range tokens.tokens {
		tokenList = append(tokenList, t)
	}

	jsonData, err := json.Marshal(tokenList)
	if err != nil {
		return fmt.Errorf("unable to marshal token database to JSON: %w", err)
	}

	folder := path.Dir(tokens.tokensFile)
	if !util.FolderExists(folder) {
		err = os.MkdirAll(folder, os.ModePerm)
		if err != nil {
			return fmt.Errorf("unable to create folder for token database: %w", err)
		}
	}

	err = ioutil.WriteFile(tokens.tokensFile, jsonData, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to write token database to file %s: %w",
			tokens.tokensFile, err)
	}

	return nil
}

func (tokens *jsonTokens) NoUserMode() bool {
	return false
}

func (tokens *jsonTokens) GetTokens(userId string) ([]Token, error) {
	tokens.mutex.Lock()
	defer tokens.mutex.Unlock()

	tokenList := make([]Token, 0)
	for _, t := range tokens.tokens {
		if t.UserId == userId {
			tokenList = append(tokenList, t.Token)
		}
	}

	return tokenList, nil
}

func (tokens *jsonTokens) CreateToken(userId string, roles *Roles) (*Token, error) {
	tokens.mutex.Lock()
	defer tokens.mutex.Unlock()

	tokenId := util.GenerateAPIToken()
	if _, found := tokens.tokens[tokenId]; found {
		return nil, fmt.Errorf("collision while generating new token %s", tokenId)
	}

	token := jsonToken{
		UserId: userId,
		Token: Token{
			Id:    tokenId,
			Roles: *roles,
		},
	}

	tokens.tokens[tokenId] = token
	err := tokens.saveTokens()
	if err != nil {
		return nil, fmt.Errorf("unable to save JSON token database: %w", err)
	}

	// Return a safe copy
	copy := token.Token
	return &copy, nil
}

func (tokens *jsonTokens) DeleteToken(tokenId string) error {
	tokens.mutex.Lock()
	defer tokens.mutex.Unlock()

	if _, found := tokens.tokens[tokenId]; !found {
		return fmt.Errorf("token %s does not exist in database", tokenId)
	}

	delete(tokens.tokens, tokenId)
	err := tokens.saveTokens()
	if err != nil {
		return fmt.Errorf("unable to save JSON token database: %w", err)
	}

	return nil
}

const ReaderRole = "READER"
const WriterRole = "WRITER"
const AdminRole = "ADMIN"

func (tokens *jsonTokens) checkToken(tokenId, role string) bool {
	tokens.mutex.Lock()
	defer tokens.mutex.Unlock()

	if _, found := tokens.tokens[tokenId]; !found {
		return false
	}
	token := tokens.tokens[tokenId]

	// Check token roles
	if role == ReaderRole && !token.Roles.Reader {
		return false
	}
	if role == WriterRole && !token.Roles.Writer {
		return false
	}
	if role == AdminRole && !token.Roles.Admin {
		return false
	}

	user, err := tokens.users.GetUser(token.UserId)
	if err != nil {
		return false
	}

	// Check user roles
	if role == ReaderRole && !user.Roles.Reader {
		return false
	}
	if role == WriterRole && !user.Roles.Writer {
		return false
	}
	if role == AdminRole && !user.Roles.Admin {
		return false
	}

	return true
}

func (tokens *jsonTokens) CanRead(tokenId string) bool {
	return tokens.checkToken(tokenId, ReaderRole)
}

func (tokens *jsonTokens) CanWrite(tokenId string) bool {
	return tokens.checkToken(tokenId, WriterRole)
}

func (tokens *jsonTokens) IsAdmin(tokenId string) bool {
	return tokens.checkToken(tokenId, AdminRole)
}
