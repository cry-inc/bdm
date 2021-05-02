package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/cry-inc/bdm/pkg/bdm/util"
	"golang.org/x/crypto/bcrypt"
)

type JsonUser struct {
	User
	Salt string
	Hash string
}

type JsonUserDatabase struct {
	usersFile string
	users     map[string]JsonUser
	mutex     sync.Mutex
}

func CreateJsonUserDatabase(dbfilePath string) (Users, error) {
	db := JsonUserDatabase{
		usersFile: dbfilePath,
		users:     make(map[string]JsonUser),
	}

	if !util.FileExists(db.usersFile) {
		err := db.saveUsers()
		if err != nil {
			return nil, fmt.Errorf("unable to create user database file %s: %w", dbfilePath, err)
		}
	}

	err := db.loadUsers()
	if err != nil {
		return nil, fmt.Errorf("unable to load user database: %w", err)
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
		return fmt.Errorf("error to unmarshal user database: %w", err)
	}

	db.users = make(map[string]JsonUser)
	for _, u := range users {
		db.users[u.Id] = u
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

func generateSalt() string {
	salt := make([]byte, 32)
	_, err := rand.Read(salt)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", salt)
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

	salt := generateSalt()
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

func (db *JsonUserDatabase) DeleteUser(id string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, found := db.users[id]; !found {
		return fmt.Errorf("user with ID %s does not exist in database", id)
	}

	delete(db.users, id)
	err := db.saveUsers()
	if err != nil {
		return fmt.Errorf("unable to save user database: %w", err)
	}

	return nil
}

func (db *JsonUserDatabase) SetRoles(id string, roles *Roles) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, found := db.users[id]; !found {
		return fmt.Errorf("user with ID %s does not exist in database", id)
	}

	user := db.users[id]
	user.Roles = *roles
	db.users[id] = user
	err := db.saveUsers()
	if err != nil {
		return fmt.Errorf("unable to save JSON user database: %w", err)
	}

	return nil
}

func (db *JsonUserDatabase) GetRoles(id string) (*Roles, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, found := db.users[id]; !found {
		return nil, fmt.Errorf("user with ID %s does not exist in database", id)
	}

	roles := db.users[id].Roles
	return &roles, nil
}

func (db *JsonUserDatabase) Authenticate(id, password string) bool {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, found := db.users[id]; !found {
		return false
	}

	hash, err := hex.DecodeString(db.users[id].Hash)
	if err != nil {
		return false
	}

	saltedPw := []byte(db.users[id].Salt + password)
	err = bcrypt.CompareHashAndPassword(hash, saltedPw)
	return err == nil
}

func (db *JsonUserDatabase) ChangePassword(id, password string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, found := db.users[id]; !found {
		return fmt.Errorf("user with ID %s does not exist in database", id)
	}

	user := db.users[id]
	saltedPw := []byte(user.Salt + password)
	hashBytes, err := bcrypt.GenerateFromPassword(saltedPw, bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("unable to generate password hash: %w", err)
	}

	user.Hash = fmt.Sprintf("%x", hashBytes)
	db.users[id] = user
	err = db.saveUsers()
	if err != nil {
		return fmt.Errorf("unable to save JSON user database: %w", err)
	}

	return nil
}
