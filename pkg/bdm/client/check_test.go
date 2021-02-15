package client

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"git.caputo.de/macaputo/bdm/pkg/bdm"
)

func TestCheckFiles(t *testing.T) {
	const testFolder = "../../../test/example"

	manifest, err := bdm.GenerateManifest("foo", testFolder)
	if err != nil {
		t.Fatal(err)
	}

	// Clean folder should pass check with clean enabled
	err = CheckFiles(manifest, testFolder, true)
	if err != nil {
		t.Fatal(err)
	}

	junkFile := path.Join(testFolder, "bla")
	err = ioutil.WriteFile(junkFile, []byte{123}, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(junkFile)

	// Dirty folder should pass check with clean disabled
	err = CheckFiles(manifest, testFolder, false)
	if err != nil {
		t.Fatal(err)
	}

	// Dirty folder should fail check with clean enabled
	err = CheckFiles(manifest, testFolder, true)
	if err == nil {
		t.Fatal()
	}
}
