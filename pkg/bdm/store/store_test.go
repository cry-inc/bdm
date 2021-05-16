package store

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/cry-inc/bdm/pkg/bdm"
	"github.com/cry-inc/bdm/pkg/bdm/util"
)

const storeFolder = "./store"

func TestStore(t *testing.T) {
	defer os.RemoveAll(storeFolder)
	store, err := New(storeFolder)
	util.AssertNoError(t, err)

	// Get package names from empty store
	names, err := store.GetNames()
	util.AssertNoError(t, err)
	util.Assert(t, len(names) == 0)

	// Try package that does not exist
	versions, err := store.GetVersions("doesnotexist")
	util.AssertNoError(t, err)
	util.Assert(t, len(versions) == 0)

	// Create test object #1
	objectData1 := []byte{1, 2, 3, 5, 5}
	objectSize1 := len(objectData1)
	objectHash1, err := util.HashStream(bytes.NewReader(objectData1))
	util.AssertNoError(t, err)

	// Ask for object that is not yet in the store
	_, err = store.GetObject(objectHash1)
	util.AssertError(t, err)

	// Publish object #1 to store
	buffer1 := bytes.NewReader(objectData1)
	object1, err := store.AddObject(buffer1)
	util.AssertNoError(t, err)
	util.Assert(t, object1.Hash == objectHash1)
	util.Assert(t, object1.Size == int64(objectSize1))

	// Check if object #2 was correctly added to the store
	object1, err = store.GetObject(objectHash1)
	util.AssertNoError(t, err)
	util.Assert(t, object1.Hash == objectHash1)
	util.Assert(t, object1.Size == int64(objectSize1))

	// Read object #1 from store and compare returned data
	reader, err := store.ReadObject(objectHash1)
	util.AssertNoError(t, err)
	readData, err := ioutil.ReadAll(reader)
	util.AssertNoError(t, err)
	util.Assert(t, reflect.DeepEqual(readData, objectData1))

	// Create valid unpublished manifest #1 with object #1
	manifest1 := bdm.Manifest{
		ManifestVersion: 1,
		PackageName:     "foo",
		Files: []bdm.File{
			{
				Path: "folder/file",
				Object: bdm.Object{
					Hash: objectHash1,
					Size: int64(objectSize1),
				},
			},
		},
	}
	manifest1.Hash = bdm.HashManifest(&manifest1)

	// Publish manifest #1 and check assigned version number
	err = store.PublishManifest(&manifest1)
	util.AssertNoError(t, err)
	util.Assert(t, manifest1.PackageVersion == 1)

	// Check if package list contains published manifest #1
	names, err = store.GetNames()
	util.AssertNoError(t, err)
	util.Assert(t, len(names) == 1)
	util.Assert(t, names[0] == manifest1.PackageName)

	// Check if list of versions for manifest #1 is working
	versions, err = store.GetVersions(manifest1.PackageName)
	util.AssertNoError(t, err)
	util.Assert(t, len(versions) == 1)
	util.Assert(t, versions[0] == 1)

	// Read publish manifest and compare with original
	readManifest, err := store.GetManifest(manifest1.PackageName, manifest1.PackageVersion)
	util.AssertNoError(t, err)
	util.Assert(t, reflect.DeepEqual(*readManifest, manifest1))

	// Create manifest #2 as unpublished copy of manifest #1
	manifest2 := manifest1
	manifest2.PackageVersion = 0
	manifest2.Published = 0
	manifest2.Hash = bdm.HashManifest(&manifest2)

	// Try to publish copy of same content as new version, should fail
	err = store.PublishManifest(&manifest2)
	util.AssertError(t, err)
	var dupErrr DuplicatePackageError
	util.Assert(t, errors.As(err, &dupErrr))

	// Create object #2
	objectData2 := []byte{6, 7, 8}
	objectSize2 := len(objectData2)
	objectHash2, err := util.HashStream(bytes.NewReader(objectData2))
	util.AssertNoError(t, err)

	// Create manifest #3 as copy of #1 with additional object #2
	manifest3 := manifest1
	manifest3.PackageVersion = 0
	manifest3.Published = 0
	manifest3.Files = []bdm.File{
		manifest1.Files[0],
		{
			Path: "folder/subfolder/file",
			Object: bdm.Object{
				Hash: objectHash2,
				Size: int64(objectSize2),
			},
		},
	}
	manifest3.Hash = bdm.HashManifest(&manifest3)

	// Object "completeness" check should fail since object #2 ist not yet added
	util.Assert(t, !AllObjectsExist(&manifest3, store))

	// Add object #2 to store and check again
	_, err = store.AddObject(bytes.NewBuffer(objectData2))
	util.AssertNoError(t, err)
	util.Assert(t, AllObjectsExist(&manifest3, store))

	// Now we try again to publish manifest #3 as new package version #2 and it should work
	err = store.PublishManifest(&manifest3)
	util.AssertNoError(t, err)
	util.Assert(t, manifest3.PackageVersion == 2)

	// Trigger store validation, which should succeed
	stats, err := ValidateStore(store)
	util.AssertNoError(t, err)

	// Check store validatsion stats
	util.Assert(t, stats["objects"] == 2)
	util.Assert(t, stats["size"] == 8)
	util.Assert(t, stats["packages"] == 2)
}

func TestCacheStore(t *testing.T) {
	defer os.RemoveAll(storeFolder)
	store, err := New(storeFolder)
	util.AssertNoError(t, err)

	// Fake a published manifest
	manifest := bdm.Manifest{
		ManifestVersion: 1,
		PackageName:     "foo",
		PackageVersion:  1,
		Published:       123456789,
		Files: []bdm.File{
			{
				Path: "folder/file",
				Object: bdm.Object{
					Hash: "abc",
					Size: 123,
				},
			},
		},
	}
	manifest.Hash = bdm.HashManifest(&manifest)

	// A manifest that is marked as published cannot be published again!
	err = store.PublishManifest(&manifest)
	util.AssertError(t, err)

	// Add valid published manifest directly to the store without publishing works
	err = store.AddManifest(&manifest)
	util.AssertNoError(t, err)
}
