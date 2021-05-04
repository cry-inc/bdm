package server

import (
	"os"
	"testing"
)

const dbFolder = "dbfolder"

func TestUserDatabase(t *testing.T) {
	const userName = "foo@bar.com"
	const userPw = "secretpw"
	const newPw = "newsecretpw"

	// There will be an error if the folder does not exist
	_, err := CreateJsonUserDatabase(dbFolder)
	if err == nil {
		t.Fatal()
	}

	// Create new database
	err = os.MkdirAll(dbFolder, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	db, err := CreateJsonUserDatabase(dbFolder)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dbFolder)

	// New DB should be empty
	users, err := db.GetUsers()
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 0 {
		t.Fatal()
	}

	// Prepare a new user
	user := User{
		Id: userName,
		Roles: Roles{
			Reader: true,
			Writer: true,
		},
	}
	// Password not long enough
	err = db.CreateUser(user, "short")
	if err == nil {
		t.Fatal()
	}
	// Use a valid password
	err = db.CreateUser(user, userPw)
	if err != nil {
		t.Fatal(err)
	}

	// Adding the same user again leads to an error
	err = db.CreateUser(user, userPw)
	if err == nil {
		t.Fatal()
	}

	// DB is no longer empty
	users, err = db.GetUsers()
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 1 {
		t.Fatal()
	}
	if users[0].Id != userName {
		t.Fatal()
	}
	if !users[0].Roles.Reader || !users[0].Roles.Writer {
		t.Fatal()
	}

	// Check if validation is possible
	if !db.Authenticate(userName, userPw) {
		t.Fatal()
	}
	if db.Authenticate(userName, "wrongpw") {
		t.Fatal()
	}

	// Check get and set roles
	roles, err := db.GetRoles(userName)
	if err != nil {
		t.Fatal(err)
	}
	// User should have read & write permission
	if !roles.Reader || !roles.Writer {
		t.Fatal()
	}
	// Update roles
	roles.Writer = false
	roles.Reader = false
	err = db.SetRoles(userName, roles)
	if err != nil {
		t.Fatal(err)
	}
	// Get new roles
	roles, err = db.GetRoles(userName)
	if err != nil {
		t.Fatal(err)
	}
	// User should no longer have read & write permission
	if roles.Reader || roles.Writer {
		t.Fatal()
	}

	// Change password
	err = db.ChangePassword(userName, newPw)
	if err != nil {
		t.Fatal(err)
	}

	// Old PW is no longer valid
	if db.Authenticate(userName, userPw) {
		t.Fatal()
	}
	if !db.Authenticate(userName, newPw) {
		t.Fatal()
	}

	// Delete the user
	err = db.DeleteUser(userName)
	if err != nil {
		t.Fatal(err)
	}

	// New DB should be empty again
	users, err = db.GetUsers()
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 0 {
		t.Fatal()
	}

	// Deleting a non-existent user should cause an error
	err = db.DeleteUser(userName)
	if err == nil {
		t.Fatal()
	}
}

func TestTokenDatabase(t *testing.T) {
	const writeUser = "writer@foo.com"
	const readUser = "reader@foo.com"
	const password = "password"

	// Create new database
	err := os.MkdirAll(dbFolder, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	db, err := CreateJsonUserDatabase(dbFolder)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dbFolder)

	// Add new users
	user1 := User{
		Id: writeUser,
		Roles: Roles{
			Reader: true,
			Writer: true,
		},
	}
	err = db.CreateUser(user1, password)
	if err != nil {
		t.Fatal(err)
	}
	user2 := User{
		Id: readUser,
		Roles: Roles{
			Reader: true,
			Writer: false,
		},
	}
	err = db.CreateUser(user2, password)
	if err != nil {
		t.Fatal(err)
	}

	// Try to add token for invalid user
	_, err = db.CreateToken("novaliduser", ReadToken)
	if err == nil {
		t.Fatal()
	}

	// Invalid token type
	_, err = db.CreateToken(readUser, "novalidtype")
	if err == nil {
		t.Fatal()
	}

	// Add read tokens for both users
	readUserReadToken, err := db.CreateToken(readUser, ReadToken)
	if err != nil {
		t.Fatal(err)
	}
	writeUserReadToken, err := db.CreateToken(writeUser, ReadToken)
	if err != nil {
		t.Fatal(err)
	}

	// Add write tokens for both users
	readUserWriteToken, err := db.CreateToken(readUser, WriteToken)
	if err != nil {
		t.Fatal(err)
	}
	writeUserWriteToken, err := db.CreateToken(writeUser, WriteToken)
	if err != nil {
		t.Fatal(err)
	}

	// Get tokens for invalid user
	_, err = db.GetTokens("invaliduser")
	if err == nil {
		t.Fatal()
	}

	// Get user tokens
	readUserTokens, err := db.GetTokens(readUser)
	if err != nil {
		t.Fatal(err)
	}
	if len(readUserTokens) != 2 {
		t.Fatal()
	}
	if !containsToken(readUserReadToken, readUserTokens) {
		t.Fatal()
	}
	if !containsToken(readUserWriteToken, readUserTokens) {
		t.Fatal()
	}
	writeUserTokens, err := db.GetTokens(writeUser)
	if err != nil {
		t.Fatal(err)
	}
	if len(writeUserTokens) != 2 {
		t.Fatal()
	}
	if !containsToken(writeUserReadToken, writeUserTokens) {
		t.Fatal()
	}
	if !containsToken(writeUserWriteToken, writeUserTokens) {
		t.Fatal()
	}

	// Check permissions of read user
	if !db.CanRead(readUserReadToken) {
		t.Fatal()
	}
	if db.CanRead(readUserWriteToken) {
		t.Fatal()
	}
	if db.CanWrite(readUserReadToken) {
		t.Fatal()
	}
	if db.CanWrite(readUserWriteToken) {
		t.Fatal()
	}

	// Check permission of write user
	if !db.CanRead(writeUserReadToken) {
		t.Fatal()
	}
	if db.CanRead(writeUserWriteToken) {
		t.Fatal()
	}
	if db.CanWrite(writeUserReadToken) {
		t.Fatal()
	}
	if !db.CanWrite(writeUserWriteToken) {
		t.Fatal()
	}

	// Delete invalid token
	err = db.DeleteToken("invalidtoken")
	if err == nil {
		t.Fatal()
	}

	// Delete tokens
	err = db.DeleteToken(readUserReadToken)
	if err != nil {
		t.Fatal(err)
	}
	err = db.DeleteToken(writeUserReadToken)
	if err != nil {
		t.Fatal(err)
	}
	err = db.DeleteToken(readUserWriteToken)
	if err != nil {
		t.Fatal(err)
	}
	err = db.DeleteToken(writeUserWriteToken)
	if err != nil {
		t.Fatal(err)
	}
}

func containsToken(tokenId string, tokens []Token) bool {
	for _, t := range tokens {
		if t.Id == tokenId {
			return true
		}
	}
	return false
}
