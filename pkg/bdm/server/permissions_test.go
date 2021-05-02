package server

import (
	"testing"
)

func TestSimplePermissions(t *testing.T) {
	perms := SimplePermissions("read", "write")

	// Check empty tokens
	if perms.CanRead("") {
		t.Fatal()
	}
	if perms.CanWrite("") {
		t.Fatal()
	}

	// Check wrong tokens
	if perms.CanRead("wrong") {
		t.Fatal()
	}
	if perms.CanWrite("wrong") {
		t.Fatal()
	}
	if perms.CanWrite("read") {
		t.Fatal()
	}

	// Check correct tokens
	if !perms.CanRead("read") {
		t.Fatal()
	}
	if !perms.CanWrite("write") {
		t.Fatal()
	}

	// Check "inherited" permissions
	if !perms.CanRead("write") {
		t.Fatal("Write token should be also able to read!")
	}

	// Test free for all reading
	perms = SimplePermissions("", "write")
	if !perms.CanRead("") {
		t.Fatal()
	}
	if !perms.CanWrite("write") {
		t.Fatal()
	}
	if !perms.CanRead("write") {
		t.Fatal()
	}
	if perms.CanRead("wrong") {
		t.Fatal()
	}
}
