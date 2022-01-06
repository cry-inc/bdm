package util

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

// Int64FromBytes decodes a byte array into a 64 bit integer
func Int64FromBytes(bytes []byte) (int64, error) {
	if len(bytes) != 8 {
		return 0, fmt.Errorf("found %d bytes, but requires 8 bytes", len(bytes))
	}
	value, readBytes := binary.Varint(bytes)
	if readBytes <= 0 {
		return 0, fmt.Errorf("error reading int64 from bytes")
	}
	return value, nil
}

// Int64ToBytes encodes a 64 bit integer into a byte array
func Int64ToBytes(value int64) []byte {
	buffer := make([]byte, 8)
	binary.PutVarint(buffer, value)
	return buffer
}

// GenerateRandomHexString generates a random hex string with byteLength * 2 characters
func GenerateRandomHexString(byteLength uint) string {
	data := make([]byte, byteLength)
	_, err := rand.Read(data)
	if err != nil {
		panic(fmt.Errorf("failed to read random data: %w", err))
	}
	return fmt.Sprintf("%x", data)
}

// GenerateAPIToken generates a random new API token
func GenerateAPIToken() string {
	return GenerateRandomHexString(32)
}
