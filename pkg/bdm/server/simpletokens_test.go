package server

import (
	"testing"
)

func TestSimpleTokens(t *testing.T) {
	tokens := SimpleTokens("read", "write", "admin")

	// Check empty tokens
	if tokens.CanRead("") {
		t.Fatal()
	}
	if tokens.CanWrite("") {
		t.Fatal()
	}

	// Check wrong tokens
	if tokens.CanRead("wrong") {
		t.Fatal()
	}
	if tokens.CanWrite("wrong") {
		t.Fatal()
	}
	if tokens.CanWrite("read") {
		t.Fatal()
	}
	if tokens.IsAdmin("wrong") {
		t.Fatal()
	}

	// Check correct tokens
	if !tokens.CanRead("read") {
		t.Fatal()
	}
	if !tokens.CanWrite("write") {
		t.Fatal()
	}
	if !tokens.IsAdmin("admin") {
		t.Fatal()
	}

	// Check "inherited" permissions
	if !tokens.CanRead("write") {
		t.Fatal("Write token should be also able to read!")
	}
	if !tokens.CanRead("admin") {
		t.Fatal("Admin token should be also able to read!")
	}
	if !tokens.CanWrite("admin") {
		t.Fatal("Admin token should be also able to write!")
	}

	// Test free for all reading without admin permissions
	tokens = SimpleTokens("", "write", "")
	if !tokens.CanRead("") {
		t.Fatal()
	}
	if !tokens.CanWrite("write") {
		t.Fatal()
	}
	if !tokens.CanRead("write") {
		t.Fatal()
	}
	if tokens.CanRead("wrong") {
		t.Fatal()
	}
	if tokens.IsAdmin("") {
		t.Fatal()
	}

	// User mode is not supported
	if !tokens.NoUserMode() {
		t.Fatal()
	}
}
