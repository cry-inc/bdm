package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/cry-inc/bdm/pkg/bdm/util"
)

type jsonToken struct {
	Token
	UserId string
}

type jsonTokens struct {
	tokensFile     string
	guestDownload  bool
	guestUpload    bool
	tokensBySecret map[string]jsonToken
	tokensById     map[string]jsonToken
	mutex          sync.Mutex
	users          Users
}

// CreateJsonTokens returns a implementation of the Tokens interface
// that uses a simple JSON file as storage for the token database.
func CreateJsonTokens(tokensFile string, users Users, guestDownload, guestUpload bool) (Tokens, error) {
	if guestUpload && !guestDownload {
		return nil, fmt.Errorf("guest uploading without guest downloading is not supported")
	}

	tokens := jsonTokens{
		tokensFile:     tokensFile,
		guestDownload:  guestDownload,
		guestUpload:    guestUpload,
		tokensBySecret: make(map[string]jsonToken),
		tokensById:     make(map[string]jsonToken),
		users:          users,
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
	jsonData, err := os.ReadFile(tokens.tokensFile)
	if err != nil {
		return fmt.Errorf("error reading token database file %s: %w", tokens.tokensFile, err)
	}

	var tokenList []jsonToken
	err = json.Unmarshal(jsonData, &tokenList)
	if err != nil {
		return fmt.Errorf("error while unmarshalling token database: %w", err)
	}

	tokens.tokensById = make(map[string]jsonToken)
	tokens.tokensBySecret = make(map[string]jsonToken)
	for _, t := range tokenList {
		tokens.tokensById[t.Id] = t
		tokens.tokensBySecret[t.Secret] = t
	}

	return nil
}

func (tokens *jsonTokens) saveTokens() error {
	tokenList := make([]jsonToken, 0)
	for _, t := range tokens.tokensById {
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

	err = os.WriteFile(tokens.tokensFile, jsonData, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to write token database to file %s: %w",
			tokens.tokensFile, err)
	}

	return nil
}

func (tokens *jsonTokens) GetTokens(userId string) ([]Token, error) {
	tokens.mutex.Lock()
	defer tokens.mutex.Unlock()

	tokenList := make([]Token, 0)
	for _, t := range tokens.tokensById {
		if t.UserId == userId {
			tokenList = append(tokenList, t.Token)
		}
	}

	return tokenList, nil
}

func (tokens *jsonTokens) CreateToken(userId, name string, expiration time.Time, roles *Roles) (*Token, error) {
	tokens.mutex.Lock()
	defer tokens.mutex.Unlock()

	tokenId := util.GenerateAPIToken()
	if _, found := tokens.tokensById[tokenId]; found {
		return nil, fmt.Errorf("collision while generating new token ID %s", tokenId)
	}

	tokenSecret := util.GenerateAPIToken()
	if _, found := tokens.tokensBySecret[tokenSecret]; found {
		return nil, fmt.Errorf("collision while generating new token secret %s", tokenSecret)
	}

	token := jsonToken{
		UserId: userId,
		Token: Token{
			Id:         tokenId,
			Name:       name,
			Secret:     tokenSecret,
			Expiration: expiration,
			Roles:      *roles,
		},
	}

	tokens.tokensById[tokenId] = token
	tokens.tokensBySecret[tokenSecret] = token
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

	if _, found := tokens.tokensById[tokenId]; !found {
		return fmt.Errorf("token with ID %s does not exist in database", tokenId)
	}

	tobeDeleted := tokens.tokensById[tokenId]
	delete(tokens.tokensById, tobeDeleted.Id)
	delete(tokens.tokensBySecret, tobeDeleted.Secret)

	err := tokens.saveTokens()
	if err != nil {
		return fmt.Errorf("unable to save JSON token database: %w", err)
	}

	return nil
}

const readerRole = "READER"
const writerRole = "WRITER"
const adminRole = "ADMIN"

func (tokens *jsonTokens) checkToken(secret, role string) bool {
	tokens.mutex.Lock()
	defer tokens.mutex.Unlock()

	if _, found := tokens.tokensBySecret[secret]; !found {
		return false
	}
	token := tokens.tokensBySecret[secret]

	// Check expiration
	if token.Expiration.Before(time.Now()) {
		return false
	}

	// Check token roles
	if role == readerRole && !token.Roles.Reader {
		return false
	}
	if role == writerRole && !token.Roles.Writer {
		return false
	}
	if role == adminRole && !token.Roles.Admin {
		return false
	}

	user, err := tokens.users.GetUser(token.UserId)
	if err != nil {
		return false
	}

	// Check user roles
	if role == readerRole && !user.Roles.Reader {
		return false
	}
	if role == writerRole && !user.Roles.Writer {
		return false
	}
	if role == adminRole && !user.Roles.Admin {
		return false
	}

	return true
}

func (tokens *jsonTokens) CanRead(secret string) bool {
	if tokens.guestDownload {
		return true
	}
	return tokens.checkToken(secret, readerRole)
}

func (tokens *jsonTokens) CanWrite(secret string) bool {
	if tokens.guestUpload {
		return true
	}
	return tokens.checkToken(secret, writerRole)
}

func (tokens *jsonTokens) IsAdmin(secret string) bool {
	return tokens.checkToken(secret, adminRole)
}
