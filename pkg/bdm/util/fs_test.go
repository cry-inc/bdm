package util

import (
	"testing"
)

func TestFileExists(t *testing.T) {
	Assert(t, !FileExists("filedoesnotexist"))
	Assert(t, !FileExists("."))
	Assert(t, !FileExists("./"))
	Assert(t, FileExists("fs.go"))
}

func TestFolderExists(t *testing.T) {
	Assert(t, !FolderExists("folderdoesnotexist"))
	Assert(t, !FolderExists("fs.go"))
	Assert(t, FolderExists("."))
	Assert(t, FolderExists(".."))
	Assert(t, FolderExists("./"))
}
