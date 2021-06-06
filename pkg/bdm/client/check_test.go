package client

import (
	"os"
	"path"
	"testing"

	"github.com/cry-inc/bdm/pkg/bdm"
	"github.com/cry-inc/bdm/pkg/bdm/util"
)

func TestCheckFiles(t *testing.T) {
	const testFolder = "../../../test/example"

	manifest, err := bdm.GenerateManifest("foo", testFolder)
	util.AssertNoError(t, err)

	// Clean folder should pass check with clean enabled
	err = CheckFiles(manifest, testFolder, true)
	util.AssertNoError(t, err)

	junkFile := path.Join(testFolder, "bla")
	err = os.WriteFile(junkFile, []byte{123}, os.ModePerm)
	util.AssertNoError(t, err)
	defer os.Remove(junkFile)

	// Dirty folder should pass check with clean disabled
	err = CheckFiles(manifest, testFolder, false)
	util.AssertNoError(t, err)

	// Dirty folder should fail check with clean enabled
	err = CheckFiles(manifest, testFolder, true)
	util.AssertError(t, err)
}
