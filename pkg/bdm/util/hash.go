package util

import (
	"fmt"
	"hash"
	"io"
	"os"

	"github.com/zeebo/blake3"
)

// CreateHasher returns the hasher to calculate all checksum for BDM
func CreateHasher() hash.Hash {
	return blake3.New()
}

// HashStream calculates the hash for all the data from a reader
func HashStream(reader io.Reader) (string, error) {
	hasher := CreateHasher()
	_, err := io.Copy(hasher, reader)
	if err != nil {
		return "", fmt.Errorf("error copying stream to hasher: %w", err)
	}
	hash := GetHashString(hasher)
	return hash, nil
}

// HashFile calculates the hash for a complete file on disk
func HashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("error opening file %s: %w", path, err)
	}
	defer file.Close()
	return HashStream(file)
}

// GetHashString returns the current hash value as hex string
func GetHashString(hasher hash.Hash) string {
	hashSum := hasher.Sum(nil)
	return fmt.Sprintf("%x", hashSum)
}
