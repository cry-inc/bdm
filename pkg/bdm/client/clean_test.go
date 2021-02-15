package client

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"git.caputo.de/macaputo/bdm/pkg/bdm"
	"git.caputo.de/macaputo/bdm/pkg/bdm/util"
)

func TestCleanPackage(t *testing.T) {
	const testFolder = "../../../test/example"

	manifest, err := bdm.GenerateManifest("foo", testFolder)
	if err != nil {
		t.Fatal(err)
	}

	// Clean a already cleaned folder should work
	err = CleanPackage(manifest, testFolder)
	if err != nil {
		t.Fatal(err)
	}

	junkFile := path.Join(testFolder, "bla")
	err = ioutil.WriteFile(junkFile, []byte{123}, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(junkFile)

	if !util.FileExists(junkFile) {
		t.Fatal("Failed to create junk file")
	}

	junkFolder := path.Join(path.Join(testFolder, "foo"), "bar")
	err = os.MkdirAll(junkFolder, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(junkFolder)

	if !util.FolderExists(junkFolder) {
		t.Fatal("Failed to create junk folder")
	}

	// Clean folder should kill the junk
	err = CleanPackage(manifest, testFolder)
	if err != nil {
		t.Fatal(err)
	}

	if util.FileExists(junkFile) {
		t.Fatal("Cleaning did not delete the junk file!")
	}

	if util.FolderExists(junkFolder) {
		t.Fatal("Cleaning did not delete the junk folder!")
	}
}
