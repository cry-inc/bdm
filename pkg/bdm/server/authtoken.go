package server

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const defaultExpiration time.Duration = 365 * 24 * time.Hour

var secret = make([]byte, 512)

type authToken struct {
	UserId  string
	Expires time.Time
	Token   string
}

func init() {
	n, err := rand.Read(secret)
	if n != len(secret) || err != nil {
		panic(err)
	}
}

func createAuthToken(userId string, expiration time.Duration) authToken {
	expires := time.Now().Add(expiration)
	expiresStr := fmt.Sprintf("%d", expires.Unix())
	signedData := expiresStr + "." + base64.StdEncoding.EncodeToString([]byte(userId))
	signer := hmac.New(sha512.New, secret)
	n, err := signer.Write([]byte(signedData))
	if n != len(signedData) || err != nil {
		panic(err)
	}
	signature := signer.Sum(nil)
	token := signedData + "." + base64.StdEncoding.EncodeToString(signature)
	return authToken{
		UserId:  userId,
		Expires: expires,
		Token:   token,
	}
}

func readAuthToken(token string) (*authToken, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("input string is not a valid auth token")
	}

	signature, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 signature")
	}

	signedData := parts[0] + "." + parts[1]
	signer := hmac.New(sha512.New, secret)
	_, err = signer.Write([]byte(signedData))
	if err != nil {
		return nil, fmt.Errorf("failed to write data for HMAC: %w", err)
	}

	expectedSignature := signer.Sum(nil)
	if !hmac.Equal(expectedSignature, signature) {
		return nil, fmt.Errorf("detected invalid signature")
	}

	expiresUnix, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expiration date: %w", err)
	}

	currentUnix := time.Now().Unix()
	if currentUnix > expiresUnix {
		return nil, fmt.Errorf("auth token already expired")
	}

	userId, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 payload")
	}

	return &authToken{
		UserId:  string(userId),
		Expires: time.Unix(expiresUnix, 0),
		Token:   token,
	}, nil
}
