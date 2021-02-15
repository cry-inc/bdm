package util

import (
	"testing"
)

func TestFileExists(t *testing.T) {
	if false != FileExists("filedoesnotexist") {
		t.Fatal()
	}
	if false != FileExists(".") {
		t.Fatal()
	}
	if false != FileExists("./") {
		t.Fatal()
	}
	if true != FileExists("fs.go") {
		t.Fatal()
	}
}

func TestFolderExists(t *testing.T) {
	if false != FolderExists("folderdoesnotexist") {
		t.Fatal()
	}
	if true != FolderExists(".") {
		t.Fatal()
	}
	if true != FolderExists("./") {
		t.Fatal()
	}
	if false != FolderExists("fs.go") {
		t.Fatal()
	}
}
