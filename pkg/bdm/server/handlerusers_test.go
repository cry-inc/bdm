package server

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/cry-inc/bdm/pkg/bdm/util"
)

const usersFile = "./users.json"

func prepareTestUsers(t *testing.T, usersFile string) Users {
	if util.FileExists(usersFile) {
		os.Remove(usersFile)
	}
	users, err := CreateJsonUsers(usersFile)
	if err != nil {
		t.Fatal(err)
	}
	err = users.CreateUser(User{Id: "reader", Roles: Roles{Reader: true}}, "readerpassword")
	if err != nil {
		t.Fatal(err)
	}
	err = users.CreateUser(User{Id: "writer", Roles: Roles{Writer: true}}, "writerpassword")
	if err != nil {
		t.Fatal(err)
	}
	err = users.CreateUser(User{Id: "admin", Roles: Roles{Admin: true}}, "adminpassword")
	if err != nil {
		t.Fatal(err)
	}
	return users
}

func TestUsersGetHandler(t *testing.T) {
	users := prepareTestUsers(t, usersFile)
	defer os.Remove(usersFile)
	router := CreateRouter(nil, nil, users, nil)

	request := createMockedRequest("GET", "/users", nil, nil)
	response := createMockedResponse()
	router.ServeHTTP(response, request)

	if response.status != 403 {
		t.Fatal(response.status)
	}

	authUser := "writer"
	request = createMockedRequest("GET", "/users", nil, &authUser)
	response = createMockedResponse()
	router.ServeHTTP(response, request)

	if response.status != 401 {
		t.Fatal(response.status)
	}

	authUser = "admin"
	request = createMockedRequest("GET", "/users", nil, &authUser)
	response = createMockedResponse()
	router.ServeHTTP(response, request)

	if response.status != 0 {
		t.Fatal(response.status)
	}

	var userList []string
	err := json.Unmarshal(response.data, &userList)
	if err != nil {
		t.Fatal(err)
	}
	if len(userList) != 3 {
		t.Fatal()
	}
	if userList[0] != "reader" {
		t.Fatal()
	}
	if userList[1] != "writer" {
		t.Fatal()
	}
	if userList[2] != "admin" {
		t.Fatal()
	}
}

func TestUserCreateGetDelete(t *testing.T) {
	users := prepareTestUsers(t, usersFile)
	defer os.Remove(usersFile)
	router := CreateRouter(nil, nil, users, nil)

	authUser := "admin"
	body := `{"Id": "newuser", "Password": "newuserpassword"}`
	request := createMockedRequest("POST", "/users", &body, &authUser)
	response := createMockedResponse()
	router.ServeHTTP(response, request)

	if response.status != 0 {
		t.Fatal(response.status)
	}
	user, err := users.GetUser("newuser")
	if err != nil {
		t.Fatal(err)
	}
	if user.Reader || user.Writer || user.Admin {
		t.Fatal()
	}

	authUser = "admin"
	request = createMockedRequest("GET", "/users/newuser", nil, &authUser)
	response = createMockedResponse()
	router.ServeHTTP(response, request)

	if response.status != 0 {
		t.Fatal(response.status)
	}

	var parsedUser User
	err = json.Unmarshal(response.data, &parsedUser)
	if err != nil {
		t.Fatal(err)
	}

	if parsedUser.Id != "newuser" {
		t.Fatal()
	}
	if parsedUser.Reader || parsedUser.Writer || parsedUser.Admin {
		t.Fatal()
	}

	authUser = "admin"
	request = createMockedRequest("DELETE", "/users/newuser", nil, &authUser)
	response = createMockedResponse()
	router.ServeHTTP(response, request)

	if response.status != 0 {
		t.Fatal(response.status)
	}

	_, err = users.GetUser("newuser")
	if err == nil {
		t.Fatal()
	}
}

func TestUserPatchHandlers(t *testing.T) {
	users := prepareTestUsers(t, usersFile)
	defer os.Remove(usersFile)
	router := CreateRouter(nil, nil, users, nil)

	authUser := "admin"
	body := `{"Password": "newadminpassword"}`
	request := createMockedRequest("PATCH", "/users/admin/password", &body, &authUser)
	response := createMockedResponse()
	router.ServeHTTP(response, request)

	if response.status != 0 {
		t.Fatal(response.status)
	}
	if !users.Authenticate("admin", "newadminpassword") {
		t.Fatal()
	}

	body = `{"Reader": true, "Writer": true, "Admin": false}`
	request = createMockedRequest("PATCH", "/users/admin/roles", &body, &authUser)
	response = createMockedResponse()
	router.ServeHTTP(response, request)

	if response.status != 0 {
		t.Fatal(response.status)
	}
	roles, err := users.GetRoles("admin")
	if err != nil {
		t.Fatal(err)
	}
	if roles.Admin || !roles.Reader || !roles.Writer {
		t.Fatal()
	}
}
