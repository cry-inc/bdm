package server

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"strings"
)

var secret = make([]byte, 512)

func init() {
	n, err := rand.Read(secret)
	if n != len(secret) || err != nil {
		panic(err)
	}
}

func CreateAuthToken(userId string) string {
	signedData := base64.StdEncoding.EncodeToString([]byte(userId))
	signer := hmac.New(sha512.New, secret)
	n, err := signer.Write([]byte(signedData))
	if n != len(signedData) || err != nil {
		panic(err)
	}
	signature := signer.Sum(nil)
	authToken := signedData + "." + base64.StdEncoding.EncodeToString(signature)
	return authToken
}

func ReadAuthToken(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return "", fmt.Errorf("input string is not a valid auth token")
	}

	signature, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 signature")
	}

	signedData := parts[0]
	signer := hmac.New(sha512.New, secret)
	_, err = signer.Write([]byte(signedData))
	if err != nil {
		return "", fmt.Errorf("failed to write data for HMAC: %w", err)
	}

	expectedSignature := signer.Sum(nil)
	if !hmac.Equal(expectedSignature, signature) {
		return "", fmt.Errorf("detected invalid signature")
	}

	userId, err := base64.StdEncoding.DecodeString(signedData)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 payload")
	}

	return string(userId), nil
}
