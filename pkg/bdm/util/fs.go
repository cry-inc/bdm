package util

import (
	"os"
)

// FileExists will return true if path is a valid file
func FileExists(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	if stat.IsDir() {
		return false
	}
	return stat.Mode().IsRegular()
}

// FolderExists will return true if path is a valid folder
func FolderExists(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return stat.IsDir()
}
