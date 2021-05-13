package server

import (
	"os"
	"testing"
)

func TestUserDatabase(t *testing.T) {
	const userName = "foo@bar.com"
	const userPw = "secretpw"
	const newPw = "newsecretpw"
	const usersFile = "users.json"

	// Create new database
	users, err := CreateJsonUsers(usersFile)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(usersFile)

	// User management is available
	if !users.Available() {
		t.Fatal()
	}

	// New DB should be empty
	userList, err := users.GetUsers()
	if err != nil {
		t.Fatal(err)
	}
	if len(userList) != 0 {
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
	err = users.CreateUser(user, "short")
	if err == nil {
		t.Fatal()
	}
	// Use a valid password
	err = users.CreateUser(user, userPw)
	if err != nil {
		t.Fatal(err)
	}

	// Adding the same user again leads to an error
	err = users.CreateUser(user, userPw)
	if err == nil {
		t.Fatal()
	}

	// DB is no longer empty
	userList, err = users.GetUsers()
	if err != nil {
		t.Fatal(err)
	}
	if len(userList) != 1 {
		t.Fatal()
	}
	if userList[0] != userName {
		t.Fatal()
	}

	// Check if validation is possible
	if !users.Authenticate(userName, userPw) {
		t.Fatal()
	}
	if users.Authenticate(userName, "wrongpw") {
		t.Fatal()
	}

	// Check get and set roles
	roles, err := users.GetRoles(userName)
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
	err = users.SetRoles(userName, roles)
	if err != nil {
		t.Fatal(err)
	}
	// Get new roles
	roles, err = users.GetRoles(userName)
	if err != nil {
		t.Fatal(err)
	}
	// User should no longer have read & write permission
	if roles.Reader || roles.Writer {
		t.Fatal()
	}

	// Change password
	err = users.ChangePassword(userName, newPw)
	if err != nil {
		t.Fatal(err)
	}

	// Old PW is no longer valid
	if users.Authenticate(userName, userPw) {
		t.Fatal()
	}
	if !users.Authenticate(userName, newPw) {
		t.Fatal()
	}

	// Delete the user
	err = users.DeleteUser(userName)
	if err != nil {
		t.Fatal(err)
	}

	// New DB should be empty again
	userList, err = users.GetUsers()
	if err != nil {
		t.Fatal(err)
	}
	if len(userList) != 0 {
		t.Fatal()
	}

	// Deleting a non-existent user should cause an error
	err = users.DeleteUser(userName)
	if err == nil {
		t.Fatal()
	}
}
