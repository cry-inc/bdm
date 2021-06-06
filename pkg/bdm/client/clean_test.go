package client

import (
	"os"
	"path"
	"testing"

	"github.com/cry-inc/bdm/pkg/bdm"
	"github.com/cry-inc/bdm/pkg/bdm/util"
)

func TestCleanPackage(t *testing.T) {
	const testFolder = "../../../test/example"

	manifest, err := bdm.GenerateManifest("foo", testFolder)
	util.AssertNoError(t, err)

	// Clean a already cleaned folder should work
	err = CleanPackage(manifest, testFolder)
	util.AssertNoError(t, err)

	junkFile := path.Join(testFolder, "bla")
	err = os.WriteFile(junkFile, []byte{123}, os.ModePerm)
	util.AssertNoError(t, err)
	defer os.Remove(junkFile)
	util.Assert(t, util.FileExists(junkFile))

	junkFolder := path.Join(path.Join(testFolder, "foo"), "bar")
	err = os.MkdirAll(junkFolder, os.ModePerm)
	util.AssertNoError(t, err)
	defer os.RemoveAll(junkFolder)
	util.Assert(t, util.FolderExists(junkFolder))

	// Clean folder should kill the junk
	err = CleanPackage(manifest, testFolder)
	util.AssertNoError(t, err)
	util.Assert(t, !util.FileExists(junkFile))
	util.Assert(t, !util.FolderExists(junkFolder))
}
