package server

import (
	"os"
	"testing"
)

const dbFile = "users.json"
const userName = "foo@bar.com"
const userPw = "secretpw"

func TestJsonUserDatabase(t *testing.T) {
	// Create new database
	db, err := CreateJsonUserDb(dbFile)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(dbFile)

	// New DB should be empty
	users, err := db.ListUsers()
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 0 {
		t.Fatal()
	}

	// Add a new user
	user := User{
		Id: userName,
		Roles: Roles{
			Reader: true,
			Writer: true,
		},
	}
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
	users, err = db.ListUsers()
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

	// Delete the user
	err = db.DeleteUser(userName)
	if err != nil {
		t.Fatal(err)
	}

	// New DB should be empty again
	users, err = db.ListUsers()
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
