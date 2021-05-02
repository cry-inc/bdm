package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cry-inc/bdm/pkg/bdm/util"
	"golang.org/x/crypto/bcrypt"
)

type JsonUser struct {
	User
	Salt string
	Hash string
}

type JsonUsersDatabase struct {
	jsonFilePath string
}

func CreateJsonUserDb(dbfilePath string) (Users, error) {
	if !util.FileExists(dbfilePath) {
		data := []byte("[]")
		err := ioutil.WriteFile(dbfilePath, data, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("unable to create user database file %s: %w", dbfilePath, err)
		}
	}
	db := JsonUsersDatabase{jsonFilePath: dbfilePath}
	return &db, nil
}

func (db *JsonUsersDatabase) loadUsers() ([]JsonUser, error) {
	jsonData, err := ioutil.ReadFile(db.jsonFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading user database file %s: %w", db.jsonFilePath, err)
	}

	var jsonUsers []JsonUser
	err = json.Unmarshal(jsonData, &jsonUsers)
	if err != nil {
		return nil, fmt.Errorf("error to unmarshal user database: %w", err)
	}

	return jsonUsers, nil
}

func (db *JsonUsersDatabase) saveUsers(users []JsonUser) error {
	jsonData, err := json.Marshal(users)
	if err != nil {
		return fmt.Errorf("unable to marshal user database to JSON: %w", err)
	}

	err = ioutil.WriteFile(db.jsonFilePath, jsonData, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to write user database to file %s: %w",
			db.jsonFilePath, err)
	}

	return nil
}

func (db *JsonUsersDatabase) findUser(id string) (*JsonUser, error) {
	users, err := db.loadUsers()
	if err != nil {
		return nil, fmt.Errorf("unable to load JSON user database: %w", err)
	}

	var user *JsonUser = nil
	for _, u := range users {
		if u.Id == id {
			user = &u
		}
	}

	return user, nil
}

func generateSalt() string {
	salt := make([]byte, 32)
	_, err := rand.Read(salt)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", salt)
}

func (db *JsonUsersDatabase) ListUsers() ([]User, error) {
	jsonUsers, err := db.loadUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to load user database: %w", err)
	}

	var users []User
	for _, u := range jsonUsers {
		users = append(users, u.User)
	}

	return users, nil
}

func (db *JsonUsersDatabase) CreateUser(user User, password string) error {
	users, err := db.loadUsers()
	if err != nil {
		return fmt.Errorf("unable to load existing user database: %w", err)
	}

	// Check for duplicate user ID
	for _, u := range users {
		if u.Id == user.Id {
			return fmt.Errorf("user ID exists already in database")
		}
	}

	salt := generateSalt()
	saltedPw := []byte(salt + password)
	hashBytes, err := bcrypt.GenerateFromPassword(saltedPw, bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("unable to generate password hash: %w", err)
	}

	hexHash := fmt.Sprintf("%x", hashBytes)
	jsonUser := JsonUser{
		User: user,
		Salt: salt,
		Hash: hexHash,
	}

	users = append(users, jsonUser)

	err = db.saveUsers(users)
	if err != nil {
		return fmt.Errorf("unable to save user database: %w", err)
	}

	return nil
}

func (db *JsonUsersDatabase) DeleteUser(id string) error {
	users, err := db.loadUsers()
	if err != nil {
		return fmt.Errorf("unable to load JSON user database: %w", err)
	}

	cleanUsers := make([]JsonUser, 0)
	for _, u := range users {
		if u.Id != id {
			cleanUsers = append(cleanUsers, u)
		}
	}

	if len(cleanUsers) != len(users)-1 {
		return fmt.Errorf("unable to find user %s in database", id)
	}

	db.saveUsers(cleanUsers)
	return nil
}

func (db *JsonUsersDatabase) SetRoles(id string, roles *Roles) error {
	users, err := db.loadUsers()
	if err != nil {
		return fmt.Errorf("unable to load JSON user database: %w", err)
	}

	for i := range users {
		if users[i].Id == id {
			users[i].Roles = *roles
			err = db.saveUsers(users)
			if err != nil {
				return fmt.Errorf("unable to save JSON user database: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("unable to find user %s in database", id)
}

func (db *JsonUsersDatabase) GetRoles(id string) (*Roles, error) {
	user, err := db.findUser(id)
	if err != nil {
		return nil, fmt.Errorf("unable to find user: %w", err)
	}

	return &user.Roles, nil
}

func (db *JsonUsersDatabase) Authenticate(id, password string) bool {
	user, err := db.findUser(id)
	if err != nil {
		return false
	}

	hash, err := hex.DecodeString(user.Hash)
	if err != nil {
		return false
	}

	saltedPw := []byte(user.Salt + password)
	err = bcrypt.CompareHashAndPassword(hash, saltedPw)
	return err == nil
}
