package server

import (
	"os"
	"testing"
)

func TestTokenDatabase(t *testing.T) {
	const writeUser = "writer@foo.com"
	const readUser = "reader@foo.com"
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
			Writer: false,
		},
	}
	err = users.CreateUser(user2, password)
	if err != nil {
		t.Fatal(err)
	}

	// Add read tokens for both users
	readUserReadToken, err := tokens.CreateToken(readUser, &Roles{Reader: true, Writer: false})
	if err != nil {
		t.Fatal(err)
	}
	writeUserReadToken, err := tokens.CreateToken(writeUser, &Roles{Reader: true, Writer: false})
	if err != nil {
		t.Fatal(err)
	}

	// Add write tokens for both users
	readUserWriteToken, err := tokens.CreateToken(readUser, &Roles{Reader: false, Writer: true})
	if err != nil {
		t.Fatal(err)
	}
	writeUserWriteToken, err := tokens.CreateToken(writeUser, &Roles{Reader: false, Writer: true})
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
	if err != nil {
		t.Fatal(err)
	}
	if len(readUserTokens) != 2 {
		t.Fatal()
	}
	if !containsToken(readUserReadToken.Id, readUserTokens) {
		t.Fatal()
	}
	if !containsToken(readUserWriteToken.Id, readUserTokens) {
		t.Fatal()
	}
	writeUserTokens, err := tokens.GetTokens(writeUser)
	if err != nil {
		t.Fatal(err)
	}
	if len(writeUserTokens) != 2 {
		t.Fatal()
	}
	if !containsToken(writeUserReadToken.Id, writeUserTokens) {
		t.Fatal()
	}
	if !containsToken(writeUserWriteToken.Id, writeUserTokens) {
		t.Fatal()
	}

	// Check permissions of read user
	if !tokens.CanRead(readUserReadToken.Id) {
		t.Fatal()
	}
	if tokens.CanRead(readUserWriteToken.Id) {
		t.Fatal()
	}
	if tokens.CanWrite(readUserReadToken.Id) {
		t.Fatal()
	}
	if tokens.CanWrite(readUserWriteToken.Id) {
		t.Fatal()
	}

	// Check permission of write user
	if !tokens.CanRead(writeUserReadToken.Id) {
		t.Fatal()
	}
	if tokens.CanRead(writeUserWriteToken.Id) {
		t.Fatal()
	}
	if tokens.CanWrite(writeUserReadToken.Id) {
		t.Fatal()
	}
	if !tokens.CanWrite(writeUserWriteToken.Id) {
		t.Fatal()
	}

	// Delete invalid token
	err = tokens.DeleteToken("invalidtoken")
	if err == nil {
		t.Fatal()
	}

	// Delete tokens
	err = tokens.DeleteToken(readUserReadToken.Id)
	if err != nil {
		t.Fatal(err)
	}
	err = tokens.DeleteToken(writeUserReadToken.Id)
	if err != nil {
		t.Fatal(err)
	}
	err = tokens.DeleteToken(readUserWriteToken.Id)
	if err != nil {
		t.Fatal(err)
	}
	err = tokens.DeleteToken(writeUserWriteToken.Id)
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
