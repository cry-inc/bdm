package server

import (
	"os"
	"testing"

	"github.com/cry-inc/bdm/pkg/bdm/util"
)

func TestJsonTokens(t *testing.T) {
	const writeUser = "writer@foo.com"
	const readUser = "reader@foo.com"
	const adminUser = "admin@foo.com"
	const password = "password"
	const tokensFile = "tokens.json"
	const usersFile = "users.json"

	// Create new user database
	users, err := CreateJsonUsers(usersFile)
	util.AssertNoError(t, err)
	defer os.RemoveAll(usersFile)

	// Create new token database
	guestDownload := false
	guestUpload := false
	tokens, err := CreateJsonTokens(tokensFile, users, guestDownload, guestUpload)
	util.AssertNoError(t, err)
	defer os.RemoveAll(tokensFile)

	// Add new users
	user1 := User{
		Id: writeUser,
		Roles: Roles{
			Reader: true,
			Writer: true,
		},
	}
	util.AssertNoError(t, users.CreateUser(user1, password))
	user2 := User{
		Id: readUser,
		Roles: Roles{
			Reader: true,
		},
	}
	util.AssertNoError(t, users.CreateUser(user2, password))
	user3 := User{
		Id: adminUser,
		Roles: Roles{
			Reader: true,
			Writer: true,
			Admin:  true,
		},
	}
	util.AssertNoError(t, users.CreateUser(user3, password))

	// Add read tokens for all users
	readUserReadToken, err := tokens.CreateToken(readUser, &Roles{Reader: true})
	util.AssertNoError(t, err)
	writeUserReadToken, err := tokens.CreateToken(writeUser, &Roles{Reader: true})
	util.AssertNoError(t, err)
	adminUserReadToken, err := tokens.CreateToken(adminUser, &Roles{Reader: true})
	util.AssertNoError(t, err)

	// Add write tokens for all users
	readUserWriteToken, err := tokens.CreateToken(readUser, &Roles{Writer: true})
	util.AssertNoError(t, err)
	writeUserWriteToken, err := tokens.CreateToken(writeUser, &Roles{Writer: true})
	util.AssertNoError(t, err)
	adminUserWriteToken, err := tokens.CreateToken(adminUser, &Roles{Writer: true})
	util.AssertNoError(t, err)

	// Add admin tokens for all users
	readUserAdminToken, err := tokens.CreateToken(readUser, &Roles{Admin: true})
	util.AssertNoError(t, err)
	writeUserAdminToken, err := tokens.CreateToken(writeUser, &Roles{Admin: true})
	util.AssertNoError(t, err)
	adminUserAdminToken, err := tokens.CreateToken(adminUser, &Roles{Admin: true})
	util.AssertNoError(t, err)

	// Get tokens for unknown user
	tokenList, err := tokens.GetTokens("unknown")
	util.AssertNoError(t, err)
	util.Assert(t, len(tokenList) == 0)

	// Get user tokens
	readUserTokens, err := tokens.GetTokens(readUser)
	util.AssertNoError(t, err)
	util.Assert(t, len(readUserTokens) == 3)
	writeUserTokens, err := tokens.GetTokens(writeUser)
	util.AssertNoError(t, err)
	util.Assert(t, len(writeUserTokens) == 3)
	adminUserTokens, err := tokens.GetTokens(adminUser)
	util.AssertNoError(t, err)
	util.Assert(t, len(adminUserTokens) == 3)

	// Check all the admin users tokens
	util.Assert(t, containsToken(adminUserReadToken.Id, adminUserTokens))
	util.Assert(t, containsToken(adminUserWriteToken.Id, adminUserTokens))
	util.Assert(t, containsToken(adminUserAdminToken.Id, adminUserTokens))

	// Check permissions of read user
	util.Assert(t, tokens.CanRead(readUserReadToken.Id))
	util.Assert(t, !tokens.CanWrite(readUserWriteToken.Id))
	util.Assert(t, !tokens.IsAdmin(readUserAdminToken.Id))

	// Check permission of write user
	util.Assert(t, tokens.CanRead(writeUserReadToken.Id))
	util.Assert(t, tokens.CanWrite(writeUserWriteToken.Id))
	util.Assert(t, !tokens.IsAdmin(writeUserAdminToken.Id))

	// Check permission of admin user
	util.Assert(t, tokens.CanRead(adminUserReadToken.Id))
	util.Assert(t, tokens.CanWrite(writeUserWriteToken.Id))
	util.Assert(t, tokens.IsAdmin(adminUserAdminToken.Id))

	// Delete invalid token
	err = tokens.DeleteToken("invalidtoken")
	util.AssertError(t, err)

	// Delete some tokens
	err = tokens.DeleteToken(readUserReadToken.Id)
	util.AssertNoError(t, err)
	err = tokens.DeleteToken(readUserWriteToken.Id)
	util.AssertNoError(t, err)
	err = tokens.DeleteToken(readUserAdminToken.Id)
	util.AssertNoError(t, err)

	// Check token count again
	readUserTokens, err = tokens.GetTokens(readUser)
	util.AssertNoError(t, err)
	util.Assert(t, len(readUserTokens) == 0)
}

func containsToken(tokenId string, tokens []Token) bool {
	for _, t := range tokens {
		if t.Id == tokenId {
			return true
		}
	}
	return false
}

func TestGuestPermissions(t *testing.T) {
	const tokensFile = "tokens.json"
	const usersFile = "users.json"

	// Create new user database
	users, err := CreateJsonUsers(usersFile)
	util.AssertNoError(t, err)
	defer os.RemoveAll(usersFile)

	{
		// Create normal token database without guest permissions
		guestDownload := false
		guestUpload := false
		tokens, err := CreateJsonTokens(tokensFile, users, guestDownload, guestUpload)
		util.AssertNoError(t, err)
		defer os.RemoveAll(tokensFile)

		util.Assert(t, !tokens.CanRead(""))
		util.Assert(t, !tokens.CanWrite(""))
	}

	{
		// Create token database with guest downloading
		guestDownload := true
		guestUpload := false
		tokens, err := CreateJsonTokens(tokensFile, users, guestDownload, guestUpload)
		util.AssertNoError(t, err)
		defer os.RemoveAll(tokensFile)

		util.Assert(t, tokens.CanRead(""))
		util.Assert(t, !tokens.CanWrite(""))
	}

	{
		// Create token database with guest up- and downloading
		guestDownload := true
		guestUpload := true
		tokens, err := CreateJsonTokens(tokensFile, users, guestDownload, guestUpload)
		util.AssertNoError(t, err)
		defer os.RemoveAll(tokensFile)

		util.Assert(t, tokens.CanRead(""))
		util.Assert(t, tokens.CanWrite(""))
	}

	{
		// Create token database with forbidden guest permissions
		guestDownload := false
		guestUpload := true
		_, err := CreateJsonTokens(tokensFile, users, guestDownload, guestUpload)
		util.AssertError(t, err)
	}
}
