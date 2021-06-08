package server

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/cry-inc/bdm/pkg/bdm/util"
)

func TestTokensGetHandler(t *testing.T) {
	users := prepareTestUsers(t, "users.json")
	defer os.Remove("users.json")
	tokens, err := CreateJsonTokens("tokens.json", users, false, false)
	util.AssertNoError(t, err)
	defer os.Remove("tokens.json")

	router := CreateRouter(nil, nil, users, tokens)

	// Guest cannot view admin tokens
	request := createMockedRequest("GET", "/users/admin/tokens", nil, nil)
	response := createMockedResponse()
	router.ServeHTTP(response, request)
	util.Assert(t, response.status == 403)

	// Another user cannot view admin tokens
	authUser := "writer"
	request = createMockedRequest("GET", "/users/admin/tokens", nil, &authUser)
	response = createMockedResponse()
	router.ServeHTTP(response, request)
	util.Assert(t, response.status == 401)

	// Admin can view its own empty list of tokens
	authUser = "admin"
	request = createMockedRequest("GET", "/users/admin/tokens", nil, &authUser)
	response = createMockedResponse()
	router.ServeHTTP(response, request)
	util.Assert(t, response.status == 0)
	util.AssertEqualString(t, "[]", string(response.data))

	// Create a test admin token
	createdToken, err := tokens.CreateToken("admin", &Roles{Admin: true, Writer: true, Reader: true})
	util.AssertNoError(t, err)

	// Now we should get one token
	authUser = "admin"
	request = createMockedRequest("GET", "/users/admin/tokens", nil, &authUser)
	response = createMockedResponse()
	router.ServeHTTP(response, request)
	util.Assert(t, response.status == 0)
	var readTokens []Token
	err = json.Unmarshal(response.data, &readTokens)
	util.AssertNoError(t, err)
	util.Assert(t, len(readTokens) == 1)
	util.Assert(t, createdToken.Id == readTokens[0].Id)
	util.Assert(t, readTokens[0].Admin && readTokens[0].Writer && readTokens[0].Reader)
}

func TestTokensPostDeleteHandler(t *testing.T) {
	users := prepareTestUsers(t, "users.json")
	defer os.Remove("users.json")
	tokens, err := CreateJsonTokens("tokens.json", users, false, false)
	util.AssertNoError(t, err)
	defer os.Remove("tokens.json")

	router := CreateRouter(nil, nil, users, tokens)

	// Create admin token for admin user with admin role
	authUser := "admin"
	body := `{"Admin": true}`
	request := createMockedRequest("POST", "/users/admin/tokens", &body, &authUser)
	response := createMockedResponse()
	router.ServeHTTP(response, request)
	util.Assert(t, response.status == 0)
	var createdToken Token
	err = json.Unmarshal(response.data, &createdToken)
	util.AssertNoError(t, err)
	util.Assert(t, createdToken.Admin && !createdToken.Writer && !createdToken.Reader)

	// Token database should now contain one token
	tokenList, err := tokens.GetTokens("admin")
	util.AssertNoError(t, err)
	util.Assert(t, len(tokenList) == 1)

	// Delete token again
	authUser = "admin"
	request = createMockedRequest("DELETE", "/users/admin/tokens/"+createdToken.Id, nil, &authUser)
	response = createMockedResponse()
	router.ServeHTTP(response, request)
	util.Assert(t, response.status == 0)

	// Token database should be empty now
	tokenList, err = tokens.GetTokens("admin")
	util.AssertNoError(t, err)
	util.Assert(t, len(tokenList) == 0)
}
