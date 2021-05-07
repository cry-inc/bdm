package server

import "testing"

func TestJwtEndToEnd(t *testing.T) {
	// Generate a JWT
	token := CreateAuthToken("foo@bar.com")

	// Parse & Validate JWT token and extract user
	userId, err := ReadAuthToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if userId != "foo@bar.com" {
		t.Fatal(userId)
	}
}

func TestGetUserIdFromJwt(t *testing.T) {
	// Invalid token layout without two dots
	_, err := ReadAuthToken("abc")
	if err == nil {
		t.Fatal()
	}

	// Invalid base64 data in signature part
	_, err = ReadAuthToken("0.a.a")
	if err == nil {
		t.Fatal()
	}

	// Invalid signature
	_, err = ReadAuthToken("0.a.eyJmb28iOiAiYmFyIn0=")
	if err == nil {
		t.Fatal()
	}
}
