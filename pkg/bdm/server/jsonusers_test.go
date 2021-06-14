package server

import (
	"os"
	"testing"

	"github.com/cry-inc/bdm/pkg/bdm/util"
)

func TestCreateJsonUsers(t *testing.T) {
	const userName = "foo@bar.com"
	const userPw = "secretpw"
	const newPw = "newsecretpw"
	const usersFile = "users.json"

	// Create new database
	defer os.RemoveAll(usersFile)
	users, err := CreateJsonUsers(usersFile)
	util.AssertNoError(t, err)

	// New DB should be empty
	userList, err := users.GetUsers()
	util.AssertNoError(t, err)
	util.Assert(t, len(userList) == 0)

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
	util.AssertError(t, err)

	// Use a valid password
	err = users.CreateUser(user, userPw)
	util.AssertNoError(t, err)

	// Adding the same user again leads to an error
	err = users.CreateUser(user, userPw)
	util.AssertError(t, err)

	// DB is no longer empty
	userList, err = users.GetUsers()
	util.AssertNoError(t, err)
	util.Assert(t, len(userList) == 1)
	util.AssertEqualString(t, userName, userList[0])

	// Check if validation is possible
	util.Assert(t, users.Authenticate(userName, userPw))
	util.Assert(t, !users.Authenticate(userName, "wrongpw"))

	// Check get and set roles
	roles, err := users.GetRoles(userName)
	util.AssertNoError(t, err)

	// User should have read & write permission
	util.Assert(t, roles.Reader)
	util.Assert(t, roles.Writer)
	util.Assert(t, !roles.Admin)

	// Update roles
	roles.Writer = false
	roles.Reader = false
	roles.Admin = true
	err = users.SetRoles(userName, roles)
	util.AssertNoError(t, err)

	// Get new roles
	roles, err = users.GetRoles(userName)
	util.AssertNoError(t, err)

	// User should no longer have read & write permission
	util.Assert(t, !roles.Reader)
	util.Assert(t, !roles.Writer)
	util.Assert(t, roles.Admin)

	// Change password
	err = users.ChangePassword(userName, newPw)
	util.AssertNoError(t, err)

	// Old PW is no longer valid
	util.Assert(t, !users.Authenticate(userName, userPw))
	util.Assert(t, users.Authenticate(userName, newPw))

	// Delete the user
	err = users.DeleteUser(userName)
	util.AssertNoError(t, err)

	// New DB should be empty again
	userList, err = users.GetUsers()
	util.AssertNoError(t, err)
	util.Assert(t, len(userList) == 0)

	// Deleting a non-existent user should cause an error
	err = users.DeleteUser(userName)
	util.AssertError(t, err)
}

func TestJsonUserIds(t *testing.T) {
	const validPassword = "mySecurePassword"
	const usersFile = "users.json"

	defer os.RemoveAll(usersFile)
	users, err := CreateJsonUsers(usersFile)
	util.AssertNoError(t, err)

	util.AssertNoError(t, users.CreateUser(User{Id: "foo"}, validPassword))
	util.AssertNoError(t, users.CreateUser(User{Id: "FOO"}, validPassword))
	util.AssertNoError(t, users.CreateUser(User{Id: "abc123ABC@-._"}, validPassword))
	util.AssertNoError(t, users.CreateUser(User{Id: "abc-123@A_B_C.com"}, validPassword))
	util.AssertError(t, users.CreateUser(User{Id: "foo::bar"}, validPassword))
	util.AssertError(t, users.CreateUser(User{Id: "<foo>"}, validPassword))
}
