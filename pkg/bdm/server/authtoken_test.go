package server

import (
	"testing"
	"time"
)

func TestJwtEndToEnd(t *testing.T) {
	// Generate a JWT
	token := createAuthToken("foo@bar.com", defaultExpiration)
	if token.UserId != "foo@bar.com" {
		t.Fatal(token)
	}
	if time.Now().After(token.Expires) {
		t.Fatal(token)
	}
	if len(token.Token) == 0 {
		t.Fatal(token)
	}

	// Parse & Validate JWT token and extract user
	readToken, err := readAuthToken(token.Token)
	if err != nil {
		t.Fatal(err)
	}
	if readToken.Token != token.Token {
		t.Fatal(readToken)
	}
	if readToken.UserId != "foo@bar.com" {
		t.Fatal(readToken)
	}
	if readToken.Expires.Equal(token.Expires) {
		t.Fatal(readToken)
	}
}

func TestGetUserIdFromJwt(t *testing.T) {
	// Invalid token layout without two dots
	_, err := readAuthToken("abc")
	if err == nil {
		t.Fatal()
	}

	// Invalid base64 data in signature part
	_, err = readAuthToken("0.a.a")
	if err == nil {
		t.Fatal()
	}

	// Invalid signature
	_, err = readAuthToken("0.a.eyJmb28iOiAiYmFyIn0=")
	if err == nil {
		t.Fatal()
	}
}
