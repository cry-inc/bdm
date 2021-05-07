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

var secret = make([]byte, 512)

const expiresAfter = "24h"

func init() {
	n, err := rand.Read(secret)
	if n != len(secret) || err != nil {
		panic(err)
	}
}

func CreateAuthToken(userId string) string {
	expireDuration, err := time.ParseDuration(expiresAfter)
	if err != nil {
		panic(err)
	}
	currentTime := time.Now().Add(expireDuration)
	expireStr := fmt.Sprintf("%d", currentTime.Unix())
	signedData := expireStr + "." + base64.StdEncoding.EncodeToString([]byte(userId))
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
	if len(parts) != 3 {
		return "", fmt.Errorf("input string is not a valid auth token")
	}

	signature, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 signature")
	}

	signedData := parts[0] + "." + parts[1]
	signer := hmac.New(sha512.New, secret)
	_, err = signer.Write([]byte(signedData))
	if err != nil {
		return "", fmt.Errorf("failed to write data for HMAC: %w", err)
	}

	expectedSignature := signer.Sum(nil)
	if !hmac.Equal(expectedSignature, signature) {
		return "", fmt.Errorf("detected invalid signature")
	}

	expirationTime, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return "", fmt.Errorf("failed to parse expiration date: %w", err)
	}

	currentTime := time.Now().Unix()
	if currentTime > expirationTime {
		return "", fmt.Errorf("auth token already expired")
	}

	userId, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 payload")
	}

	return string(userId), nil
}
