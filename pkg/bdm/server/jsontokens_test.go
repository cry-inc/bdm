package server

import (
	"os"
	"testing"
)

func TestTokenDatabase(t *testing.T) {
	const writeUser = "writer@foo.com"
	const readUser = "reader@foo.com"
	const adminUser = "admin@foo.com"
	const password = "password"
	const tokensFile = "tokens.json"
	const usersFile = "users.json"

	// Create new user database
	users, err := CreateJsonUsers(usersFile)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(usersFile)

	// Create new token database
	tokens, err := CreateJsonTokens(tokensFile, users)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tokensFile)

	// Add new users
	user1 := User{
		Id: writeUser,
		Roles: Roles{
			Reader: true,
			Writer: true,
		},
	}
	err = users.CreateUser(user1, password)
	if err != nil {
		t.Fatal(err)
	}
	user2 := User{
		Id: readUser,
		Roles: Roles{
			Reader: true,
		},
	}
	err = users.CreateUser(user2, password)
	if err != nil {
		t.Fatal(err)
	}
	user3 := User{
		Id: adminUser,
		Roles: Roles{
			Reader: true,
			Writer: true,
			Admin:  true,
		},
	}
	err = users.CreateUser(user3, password)
	if err != nil {
		t.Fatal(err)
	}

	// Add read tokens for all users
	readUserReadToken, err := tokens.CreateToken(readUser, &Roles{Reader: true})
	if err != nil {
		t.Fatal(err)
	}
	writeUserReadToken, err := tokens.CreateToken(writeUser, &Roles{Reader: true})
	if err != nil {
		t.Fatal(err)
	}
	adminUserReadToken, err := tokens.CreateToken(adminUser, &Roles{Reader: true})
	if err != nil {
		t.Fatal(err)
	}

	// Add write tokens for all users
	readUserWriteToken, err := tokens.CreateToken(readUser, &Roles{Writer: true})
	if err != nil {
		t.Fatal(err)
	}
	writeUserWriteToken, err := tokens.CreateToken(writeUser, &Roles{Writer: true})
	if err != nil {
		t.Fatal(err)
	}
	adminUserWriteToken, err := tokens.CreateToken(adminUser, &Roles{Writer: true})
	if err != nil {
		t.Fatal(err)
	}

	// Add admin tokens for all users
	readUserAdminToken, err := tokens.CreateToken(readUser, &Roles{Admin: true})
	if err != nil {
		t.Fatal(err)
	}
	writeUserAdminToken, err := tokens.CreateToken(writeUser, &Roles{Admin: true})
	if err != nil {
		t.Fatal(err)
	}
	adminUserAdminToken, err := tokens.CreateToken(adminUser, &Roles{Admin: true})
	if err != nil {
		t.Fatal(err)
	}

	// Get tokens for unknown user
	tokenList, err := tokens.GetTokens("unknown")
	if err != nil {
		t.Fatal(err)
	}
	if len(tokenList) != 0 {
		t.Fatal()
	}

	// Get user tokens
	readUserTokens, err := tokens.GetTokens(readUser)
	if err != nil || len(readUserTokens) != 3 {
		t.Fatal()
	}
	writeUserTokens, err := tokens.GetTokens(writeUser)
	if err != nil || len(writeUserTokens) != 3 {
		t.Fatal()
	}
	adminUserTokens, err := tokens.GetTokens(adminUser)
	if err != nil || len(adminUserTokens) != 3 {
		t.Fatal()
	}

	// Check all the admin users tokens
	if !containsToken(adminUserReadToken.Id, adminUserTokens) {
		t.Fatal()
	}
	if !containsToken(adminUserWriteToken.Id, adminUserTokens) {
		t.Fatal()
	}
	if !containsToken(adminUserAdminToken.Id, adminUserTokens) {
		t.Fatal()
	}

	// Check permissions of read user
	if !tokens.CanRead(readUserReadToken.Id) {
		t.Fatal()
	}
	if tokens.CanWrite(readUserWriteToken.Id) {
		t.Fatal()
	}
	if tokens.IsAdmin(readUserAdminToken.Id) {
		t.Fatal()
	}

	// Check permission of write user
	if !tokens.CanRead(writeUserReadToken.Id) {
		t.Fatal()
	}
	if !tokens.CanWrite(writeUserWriteToken.Id) {
		t.Fatal()
	}
	if tokens.IsAdmin(writeUserAdminToken.Id) {
		t.Fatal()
	}

	// Check permission of admin user
	if !tokens.CanRead(adminUserReadToken.Id) {
		t.Fatal()
	}
	if !tokens.CanWrite(writeUserWriteToken.Id) {
		t.Fatal()
	}
	if !tokens.IsAdmin(adminUserAdminToken.Id) {
		t.Fatal()
	}

	// Delete invalid token
	err = tokens.DeleteToken("invalidtoken")
	if err == nil {
		t.Fatal()
	}

	// Delete some tokens
	err = tokens.DeleteToken(readUserReadToken.Id)
	if err != nil {
		t.Fatal(err)
	}
	err = tokens.DeleteToken(readUserWriteToken.Id)
	if err != nil {
		t.Fatal(err)
	}
	err = tokens.DeleteToken(readUserAdminToken.Id)
	if err != nil {
		t.Fatal(err)
	}

	// Check token count again
	readUserTokens, err = tokens.GetTokens(readUser)
	if err != nil || len(readUserTokens) != 0 {
		t.Fatal()
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
