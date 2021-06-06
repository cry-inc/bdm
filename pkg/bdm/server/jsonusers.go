package server

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"sync"

	"github.com/cry-inc/bdm/pkg/bdm/util"
	"golang.org/x/crypto/bcrypt"
)

type jsonUser struct {
	User
	Salt string
	Hash string
}

type jsonUsers struct {
	usersFile string
	users     map[string]jsonUser
	mutex     sync.Mutex
}

func CreateJsonUsers(usersFile string) (Users, error) {
	users := jsonUsers{
		usersFile: usersFile,
		users:     make(map[string]jsonUser),
	}

	if !util.FileExists(users.usersFile) {
		err := users.saveUsers()
		if err != nil {
			return nil, fmt.Errorf("unable to create user database file %s: %w", users.usersFile, err)
		}
	}

	err := users.loadUsers()
	if err != nil {
		return nil, fmt.Errorf("unable to load user database: %w", err)
	}

	return &users, nil
}

func (users *jsonUsers) loadUsers() error {
	jsonData, err := os.ReadFile(users.usersFile)
	if err != nil {
		return fmt.Errorf("error reading user database file %s: %w", users.usersFile, err)
	}

	var userList []jsonUser
	err = json.Unmarshal(jsonData, &userList)
	if err != nil {
		return fmt.Errorf("error while unmarshalling user database: %w", err)
	}

	users.users = make(map[string]jsonUser)
	for _, u := range userList {
		users.users[u.Id] = u
	}

	return nil
}

func (users *jsonUsers) saveUsers() error {
	userList := make([]jsonUser, 0)

	for _, u := range users.users {
		userList = append(userList, u)
	}

	jsonData, err := json.Marshal(userList)
	if err != nil {
		return fmt.Errorf("unable to marshal user database to JSON: %w", err)
	}

	folder := path.Dir(users.usersFile)
	if !util.FolderExists(folder) {
		err = os.MkdirAll(folder, os.ModePerm)
		if err != nil {
			return fmt.Errorf("unable to create folder for user database: %w", err)
		}
	}

	err = os.WriteFile(users.usersFile, jsonData, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to write user database to file %s: %w",
			users.usersFile, err)
	}

	return nil
}

func (users *jsonUsers) Available() bool {
	return true
}

func (users *jsonUsers) GetUsers() ([]string, error) {
	users.mutex.Lock()
	defer users.mutex.Unlock()

	var userList []string
	for _, u := range users.users {
		userList = append(userList, u.Id)
	}

	return userList, nil
}

func (users *jsonUsers) GetUser(userId string) (*User, error) {
	users.mutex.Lock()
	defer users.mutex.Unlock()

	if user, found := users.users[userId]; found {
		return &user.User, nil
	}

	return nil, fmt.Errorf("user not found in database")
}

func (users *jsonUsers) CreateUser(user User, password string) error {
	valid, err := regexp.MatchString(`^[a-zA-Z0-9_.@-]+$`, user.Id)
	if err != nil {
		return fmt.Errorf("failed to compile regex: %w", err)
	}
	if !valid {
		return fmt.Errorf("invalid user ID characters")
	}

	users.mutex.Lock()
	defer users.mutex.Unlock()

	// Check for duplicate user ID
	if _, found := users.users[user.Id]; found {
		return fmt.Errorf("user ID exists already in database")
	}

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	salt := util.GenerateRandomHexString(16)
	saltedPw := []byte(salt + password)
	hashBytes, err := bcrypt.GenerateFromPassword(saltedPw, bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("unable to generate password hash: %w", err)
	}

	hexHash := fmt.Sprintf("%x", hashBytes)
	users.users[user.Id] = jsonUser{
		User: user,
		Salt: salt,
		Hash: hexHash,
	}

	err = users.saveUsers()
	if err != nil {
		return fmt.Errorf("unable to save user database: %w", err)
	}

	return nil
}

func (users *jsonUsers) DeleteUser(userId string) error {
	users.mutex.Lock()
	defer users.mutex.Unlock()

	if _, found := users.users[userId]; !found {
		return fmt.Errorf("user with ID %s does not exist in database", userId)
	}

	delete(users.users, userId)
	err := users.saveUsers()
	if err != nil {
		return fmt.Errorf("unable to save user database: %w", err)
	}

	return nil
}

func (users *jsonUsers) SetRoles(userId string, roles *Roles) error {
	users.mutex.Lock()
	defer users.mutex.Unlock()

	if _, found := users.users[userId]; !found {
		return fmt.Errorf("user with ID %s does not exist in database", userId)
	}

	user := users.users[userId]
	user.Roles = *roles
	users.users[userId] = user
	err := users.saveUsers()
	if err != nil {
		return fmt.Errorf("unable to save JSON user database: %w", err)
	}

	return nil
}

func (users *jsonUsers) GetRoles(userId string) (*Roles, error) {
	users.mutex.Lock()
	defer users.mutex.Unlock()

	if _, found := users.users[userId]; !found {
		return nil, fmt.Errorf("user with ID %s does not exist in database", userId)
	}

	// Return safe copy
	roles := users.users[userId].Roles
	return &roles, nil
}

func (users *jsonUsers) Authenticate(userId, password string) bool {
	users.mutex.Lock()
	defer users.mutex.Unlock()

	if _, found := users.users[userId]; !found {
		return false
	}

	hash, err := hex.DecodeString(users.users[userId].Hash)
	if err != nil {
		return false
	}

	saltedPw := []byte(users.users[userId].Salt + password)
	err = bcrypt.CompareHashAndPassword(hash, saltedPw)
	return err == nil
}

func (users *jsonUsers) ChangePassword(userId, password string) error {
	users.mutex.Lock()
	defer users.mutex.Unlock()

	if _, found := users.users[userId]; !found {
		return fmt.Errorf("user with ID %s does not exist in database", userId)
	}

	user := users.users[userId]
	saltedPw := []byte(user.Salt + password)
	hashBytes, err := bcrypt.GenerateFromPassword(saltedPw, bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("unable to generate password hash: %w", err)
	}

	user.Hash = fmt.Sprintf("%x", hashBytes)
	users.users[userId] = user
	err = users.saveUsers()
	if err != nil {
		return fmt.Errorf("unable to save JSON user database: %w", err)
	}

	return nil
}
