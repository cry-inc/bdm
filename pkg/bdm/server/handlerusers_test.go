package server

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/cry-inc/bdm/pkg/bdm/util"
)

func hasUser(users []*User, userId string) bool {
	found := false
	for _, u := range users {
		if u.Id == userId {
			found = true
			break
		}
	}
	return found
}

func TestUsersGetHandler(t *testing.T) {
	users := prepareTestUsers(t, "users.json")
	defer os.Remove("users.json")
	router := CreateRouter(nil, nil, users, nil)

	request := createMockedRequest("GET", "/users", nil, nil)
	response := createMockedResponse()
	router.ServeHTTP(response, request)
	util.Assert(t, response.status == 403)

	authUser := "writer"
	request = createMockedRequest("GET", "/users", nil, &authUser)
	response = createMockedResponse()
	router.ServeHTTP(response, request)
	util.Assert(t, response.status == 401)

	authUser = "admin"
	request = createMockedRequest("GET", "/users", nil, &authUser)
	response = createMockedResponse()
	router.ServeHTTP(response, request)
	util.Assert(t, response.status == 0)

	var userList []*User
	err := json.Unmarshal(response.data, &userList)
	util.AssertNoError(t, err)
	util.Assert(t, len(userList) == 3)
	util.Assert(t, hasUser(userList, "reader"))
	util.Assert(t, hasUser(userList, "writer"))
	util.Assert(t, hasUser(userList, "admin"))
}

func TestUserCreateGetDelete(t *testing.T) {
	users := prepareTestUsers(t, "users.json")
	defer os.Remove("users.json")
	router := CreateRouter(nil, nil, users, nil)

	authUser := "admin"
	body := `{"Id": "newuser", "Password": "newuserpassword"}`
	request := createMockedRequest("POST", "/users", &body, &authUser)
	response := createMockedResponse()
	router.ServeHTTP(response, request)
	util.Assert(t, response.status == 0)

	user, err := users.GetUser("newuser")
	util.AssertNoError(t, err)
	util.Assert(t, !user.Reader)
	util.Assert(t, !user.Writer)
	util.Assert(t, !user.Admin)

	authUser = "admin"
	request = createMockedRequest("GET", "/users/newuser", nil, &authUser)
	response = createMockedResponse()
	router.ServeHTTP(response, request)
	util.Assert(t, response.status == 0)

	var parsedUser User
	err = json.Unmarshal(response.data, &parsedUser)
	util.AssertNoError(t, err)
	util.Assert(t, parsedUser.Id == "newuser")
	util.Assert(t, !parsedUser.Reader)
	util.Assert(t, !parsedUser.Writer)
	util.Assert(t, !parsedUser.Admin)

	authUser = "admin"
	request = createMockedRequest("DELETE", "/users/newuser", nil, &authUser)
	response = createMockedResponse()
	router.ServeHTTP(response, request)
	util.Assert(t, response.status == 0)

	_, err = users.GetUser("newuser")
	util.AssertError(t, err)
}

func TestUserPatchHandlers(t *testing.T) {
	users := prepareTestUsers(t, "users.json")
	defer os.Remove("users.json")
	router := CreateRouter(nil, nil, users, nil)

	authUser := "admin"
	body := `{"Password": "newadminpassword"}`
	request := createMockedRequest("PATCH", "/users/admin/password", &body, &authUser)
	response := createMockedResponse()
	router.ServeHTTP(response, request)
	util.Assert(t, response.status == 0)
	util.Assert(t, users.Authenticate("admin", "newadminpassword"))

	body = `{"Reader": true, "Writer": true, "Admin": false}`
	request = createMockedRequest("PATCH", "/users/admin/roles", &body, &authUser)
	response = createMockedResponse()
	router.ServeHTTP(response, request)
	util.Assert(t, response.status == 0)
	roles, err := users.GetRoles("admin")
	util.AssertNoError(t, err)
	util.Assert(t, roles.Reader)
	util.Assert(t, roles.Writer)
	util.Assert(t, !roles.Admin)
}
