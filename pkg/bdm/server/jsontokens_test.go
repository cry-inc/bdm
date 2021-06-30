package server

import (
	"os"
	"testing"
	"time"

	"github.com/cry-inc/bdm/pkg/bdm/util"
)

func containsToken(tokenId string, tokens []Token) bool {
	for _, t := range tokens {
		if t.Id == tokenId {
			return true
		}
	}
	return false
}

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
	util.AssertNoError(t, users.CreateUser(User{
		Id: writeUser,
		Roles: Roles{
			Reader: true,
			Writer: true,
		},
	}, password))
	util.AssertNoError(t, users.CreateUser(User{
		Id: readUser,
		Roles: Roles{
			Reader: true,
		},
	}, password))
	util.AssertNoError(t, users.CreateUser(User{
		Id: adminUser,
		Roles: Roles{
			Reader: true,
			Writer: true,
			Admin:  true,
		},
	}, password))

	expire := time.Now().Add(time.Hour)

	// Add read tokens for all users
	readUserReadToken, err := tokens.CreateToken(readUser, "ReadRead", expire, &Roles{Reader: true})
	util.AssertNoError(t, err)
	writeUserReadToken, err := tokens.CreateToken(writeUser, "WriteRead", expire, &Roles{Reader: true})
	util.AssertNoError(t, err)
	adminUserReadToken, err := tokens.CreateToken(adminUser, "AdminRead", expire, &Roles{Reader: true})
	util.AssertNoError(t, err)

	// Add write tokens for all users
	readUserWriteToken, err := tokens.CreateToken(readUser, "ReadWrite", expire, &Roles{Writer: true})
	util.AssertNoError(t, err)
	writeUserWriteToken, err := tokens.CreateToken(writeUser, "WriteWrite", expire, &Roles{Writer: true})
	util.AssertNoError(t, err)
	adminUserWriteToken, err := tokens.CreateToken(adminUser, "AdminWrite", expire, &Roles{Writer: true})
	util.AssertNoError(t, err)

	// Add admin tokens for all users
	readUserAdminToken, err := tokens.CreateToken(readUser, "ReadAdmin", expire, &Roles{Admin: true})
	util.AssertNoError(t, err)
	writeUserAdminToken, err := tokens.CreateToken(writeUser, "WriteAdmin", expire, &Roles{Admin: true})
	util.AssertNoError(t, err)
	adminUserAdminToken, err := tokens.CreateToken(adminUser, "AdminAdmin", expire, &Roles{Admin: true})
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
	util.Assert(t, tokens.CanRead(readUserReadToken.Secret))
	util.Assert(t, !tokens.CanWrite(readUserWriteToken.Secret))
	util.Assert(t, !tokens.IsAdmin(readUserAdminToken.Secret))

	// Check permission of write user
	util.Assert(t, tokens.CanRead(writeUserReadToken.Secret))
	util.Assert(t, tokens.CanWrite(writeUserWriteToken.Secret))
	util.Assert(t, !tokens.IsAdmin(writeUserAdminToken.Secret))

	// Check permission of admin user
	util.Assert(t, tokens.CanRead(adminUserReadToken.Secret))
	util.Assert(t, tokens.CanWrite(writeUserWriteToken.Secret))
	util.Assert(t, tokens.IsAdmin(adminUserAdminToken.Secret))

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

func TestTokenExpiration(t *testing.T) {
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
	util.AssertNoError(t, users.CreateUser(User{
		Id: "user",
		Roles: Roles{
			Reader: true,
			Writer: true,
			Admin:  true,
		},
	}, "password"))

	// Create an expired token
	expiration := time.Now().Add(-time.Hour)
	expiredToken, err := tokens.CreateToken("user", "token", expiration, &Roles{Reader: true})
	util.AssertNoError(t, err)
	util.Assert(t, !tokens.CanRead(expiredToken.Secret))

	// Create an valid token
	expiration = time.Now().Add(time.Hour)
	validToken, err := tokens.CreateToken("user", "token", expiration, &Roles{Reader: true})
	util.AssertNoError(t, err)
	util.Assert(t, tokens.CanRead(validToken.Secret))
}
