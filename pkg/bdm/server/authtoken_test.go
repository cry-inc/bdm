package server

import (
	"testing"
	"time"

	"github.com/cry-inc/bdm/pkg/bdm/util"
)

func TestJwtEndToEnd(t *testing.T) {
	// Generate a JWT
	token := createAuthToken("foo@bar.com", defaultExpiration)
	util.AssertEqualString(t, "foo@bar.com", token.UserId)
	util.Assert(t, time.Now().Before(token.Expires))
	util.Assert(t, len(token.Token) > 0)

	// Parse & Validate JWT token and extract user
	readToken, err := readAuthToken(token.Token)
	util.AssertNoError(t, err)
	util.Assert(t, readToken.Token == token.Token)
	util.AssertEqualString(t, "foo@bar.com", readToken.UserId)
	util.Assert(t, readToken.Expires.Unix() == token.Expires.Unix())

	// Test expired token
	token = createAuthToken("foo@bar.com", -10*time.Second)
	_, err = readAuthToken(token.Token)
	util.AssertError(t, err)
}

func TestGetUserIdFromJwt(t *testing.T) {
	// Invalid token layout without two dots
	_, err := readAuthToken("abc")
	util.AssertError(t, err)

	// Invalid base64 data in signature part
	_, err = readAuthToken("0.a.a")
	util.AssertError(t, err)

	// Invalid signature
	_, err = readAuthToken("0.a.eyJmb28iOiAiYmFyIn0=")
	util.AssertError(t, err)
}
