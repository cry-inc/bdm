package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"github.com/cry-inc/bdm/pkg/bdm/util"
	"golang.org/x/crypto/bcrypt"
)

type JsonUser struct {
	User
	Salt string
	Hash string
}

type JsonToken struct {
	Token
	UserId string
}

type JsonUserDatabase struct {
	usersFile  string
	tokensFile string
	users      map[string]JsonUser
	tokens     map[string]JsonToken
	mutex      sync.Mutex
}

const usersFileName = "users.json"
const tokensFileName = "tokens.json"

func CreateJsonUserDatabase(dbFolder string) (Users, error) {
	db := JsonUserDatabase{
		usersFile:  path.Join(dbFolder, usersFileName),
		tokensFile: path.Join(dbFolder, tokensFileName),
		users:      make(map[string]JsonUser),
		tokens:     make(map[string]JsonToken),
	}

	if !util.FolderExists(dbFolder) {
		return nil, fmt.Errorf("folder %s does not exist", dbFolder)
	}

	if !util.FileExists(db.usersFile) {
		err := db.saveUsers()
		if err != nil {
			return nil, fmt.Errorf("unable to create user database file %s: %w", db.usersFile, err)
		}
	}

	if !util.FileExists(db.tokensFile) {
		err := db.saveTokens()
		if err != nil {
			return nil, fmt.Errorf("unable to create user database file %s: %w", db.tokensFile, err)
		}
	}

	err := db.loadUsers()
	if err != nil {
		return nil, fmt.Errorf("unable to load user database: %w", err)
	}

	err = db.loadTokens()
	if err != nil {
		return nil, fmt.Errorf("unable to load token database: %w", err)
	}

	return &db, nil
}

func (db *JsonUserDatabase) loadUsers() error {
	jsonData, err := ioutil.ReadFile(db.usersFile)
	if err != nil {
		return fmt.Errorf("error reading user database file %s: %w", db.usersFile, err)
	}

	var users []JsonUser
	err = json.Unmarshal(jsonData, &users)
	if err != nil {
		return fmt.Errorf("error while unmarshalling user database: %w", err)
	}

	db.users = make(map[string]JsonUser)
	for _, u := range users {
		db.users[u.Id] = u
	}

	return nil
}

func (db *JsonUserDatabase) loadTokens() error {
	jsonData, err := ioutil.ReadFile(db.tokensFile)
	if err != nil {
		return fmt.Errorf("error reading token database file %s: %w", db.tokensFile, err)
	}

	var tokens []JsonToken
	err = json.Unmarshal(jsonData, &tokens)
	if err != nil {
		return fmt.Errorf("error while unmarshalling token database: %w", err)
	}

	db.tokens = make(map[string]JsonToken)
	for _, t := range tokens {
		db.tokens[t.Id] = t
	}

	return nil
}

func (db *JsonUserDatabase) saveUsers() error {
	users := make([]JsonUser, 0)

	for _, u := range db.users {
		users = append(users, u)
	}

	jsonData, err := json.Marshal(users)
	if err != nil {
		return fmt.Errorf("unable to marshal user database to JSON: %w", err)
	}

	err = ioutil.WriteFile(db.usersFile, jsonData, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to write user database to file %s: %w",
			db.usersFile, err)
	}

	return nil
}

func (db *JsonUserDatabase) saveTokens() error {
	tokens := make([]JsonToken, 0)

	for _, t := range db.tokens {
		tokens = append(tokens, t)
	}

	jsonData, err := json.Marshal(tokens)
	if err != nil {
		return fmt.Errorf("unable to marshal token database to JSON: %w", err)
	}

	err = ioutil.WriteFile(db.tokensFile, jsonData, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to write token database to file %s: %w",
			db.usersFile, err)
	}

	return nil
}

func generateRandomHexString(byteLength uint) string {
	data := make([]byte, byteLength)
	_, err := rand.Read(data)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", data)
}

func (db *JsonUserDatabase) ListUsers() ([]User, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	var users []User
	for _, u := range db.users {
		users = append(users, u.User)
	}

	return users, nil
}

func (db *JsonUserDatabase) CreateUser(user User, password string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	// Check for duplicate user ID
	if _, found := db.users[user.Id]; found {
		return fmt.Errorf("user ID exists already in database")
	}

	salt := generateRandomHexString(16)
	saltedPw := []byte(salt + password)
	hashBytes, err := bcrypt.GenerateFromPassword(saltedPw, bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("unable to generate password hash: %w", err)
	}

	hexHash := fmt.Sprintf("%x", hashBytes)
	db.users[user.Id] = JsonUser{
		User: user,
		Salt: salt,
		Hash: hexHash,
	}

	err = db.saveUsers()
	if err != nil {
		return fmt.Errorf("unable to save user database: %w", err)
	}

	return nil
}

func (db *JsonUserDatabase) DeleteUser(userId string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, found := db.users[userId]; !found {
		return fmt.Errorf("user with ID %s does not exist in database", userId)
	}

	delete(db.users, userId)
	err := db.saveUsers()
	if err != nil {
		return fmt.Errorf("unable to save user database: %w", err)
	}

	return nil
}

func (db *JsonUserDatabase) SetRoles(userId string, roles *Roles) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, found := db.users[userId]; !found {
		return fmt.Errorf("user with ID %s does not exist in database", userId)
	}

	user := db.users[userId]
	user.Roles = *roles
	db.users[userId] = user
	err := db.saveUsers()
	if err != nil {
		return fmt.Errorf("unable to save JSON user database: %w", err)
	}

	return nil
}

func (db *JsonUserDatabase) GetRoles(userId string) (*Roles, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, found := db.users[userId]; !found {
		return nil, fmt.Errorf("user with ID %s does not exist in database", userId)
	}

	roles := db.users[userId].Roles
	return &roles, nil
}

func (db *JsonUserDatabase) Authenticate(userId, password string) bool {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, found := db.users[userId]; !found {
		return false
	}

	hash, err := hex.DecodeString(db.users[userId].Hash)
	if err != nil {
		return false
	}

	saltedPw := []byte(db.users[userId].Salt + password)
	err = bcrypt.CompareHashAndPassword(hash, saltedPw)
	return err == nil
}

func (db *JsonUserDatabase) ChangePassword(userId, password string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, found := db.users[userId]; !found {
		return fmt.Errorf("user with ID %s does not exist in database", userId)
	}

	user := db.users[userId]
	saltedPw := []byte(user.Salt + password)
	hashBytes, err := bcrypt.GenerateFromPassword(saltedPw, bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("unable to generate password hash: %w", err)
	}

	user.Hash = fmt.Sprintf("%x", hashBytes)
	db.users[userId] = user
	err = db.saveUsers()
	if err != nil {
		return fmt.Errorf("unable to save JSON user database: %w", err)
	}

	return nil
}

func (db *JsonUserDatabase) GetTokens(userId string) ([]Token, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, found := db.users[userId]; !found {
		return nil, fmt.Errorf("user with ID %s does not exist in database", userId)
	}

	tokens := make([]Token, 0)
	for _, t := range db.tokens {
		if t.UserId == userId {
			tokens = append(tokens, t.Token)
		}
	}

	return tokens, nil
}

func (db *JsonUserDatabase) AddToken(userId, tokenType string) (string, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, found := db.users[userId]; !found {
		return "", fmt.Errorf("user with ID %s does not exist in database", userId)
	}

	token := generateRandomHexString(16)
	if _, found := db.tokens[token]; found {
		return "", fmt.Errorf("collision while generating new token %s", token)
	}

	jsonToken := JsonToken{
		UserId: userId,
		Token: Token{
			Id:   token,
			Type: tokenType,
		},
	}

	db.tokens[token] = jsonToken
	err := db.saveTokens()
	if err != nil {
		return "", fmt.Errorf("unable to save JSON token database: %w", err)
	}

	return token, nil
}

func (db *JsonUserDatabase) RemoveToken(tokenId string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, found := db.tokens[tokenId]; !found {
		return fmt.Errorf("token %s does not exist in database", tokenId)
	}

	delete(db.tokens, tokenId)
	err := db.saveTokens()
	if err != nil {
		return fmt.Errorf("unable to save JSON token database: %w", err)
	}

	return nil
}
