package bdm

import (
	"testing"
)

func TestDefaultLimits(t *testing.T) {
	limits := ManifestLimits{}
	manifest := Manifest{}

	// Check empty manifest with default empty limits
	err := CheckManifestLimits(&manifest, &limits)
	if err != nil {
		t.Fatal(err)
	}

	// Add file to manifest
	manifest.Files = []File{
		{
			Path:   "path/to/file",
			Object: Object{Size: 123, Hash: "abc"},
		},
	}

	// Check non-empty manifest with default limits
	err = CheckManifestLimits(&manifest, &limits)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCustomLimits(t *testing.T) {
	// Define some limits for testing
	limits := ManifestLimits{
		MaxFileSize:   1000,
		MaxFilesCount: 2,
		MaxSize:       1500,
		MaxPathLength: 10,
	}

	// Prepare a small manifest within limits to check
	manifest := Manifest{}
	manifest.Files = []File{
		{
			Path:   "file1",
			Object: Object{Size: 1000, Hash: "abc"},
		},
		{
			Path:   "file2",
			Object: Object{Size: 200, Hash: "def"},
		},
	}

	// Check manifest with limits
	err := CheckManifestLimits(&manifest, &limits)
	if err != nil {
		t.Fatal(err)
	}

	// Check invalid file size
	manifest.Files[0].Object.Size = 1001
	err = CheckManifestLimits(&manifest, &limits)
	if err == nil {
		t.Fatal()
	}
	manifest.Files[0].Object.Size = 1000

	// Check invalid size
	manifest.Files[1].Object.Size = 1000
	err = CheckManifestLimits(&manifest, &limits)
	if err == nil {
		t.Fatal()
	}
	manifest.Files[1].Object.Size = 200

	// Check invalid path length
	manifest.Files[0].Path = "file/path/is/to/long.txt"
	err = CheckManifestLimits(&manifest, &limits)
	if err == nil {
		t.Fatal()
	}
	manifest.Files[0].Path = "file1"

	// Check invalid file count
	newFile := File{Path: "file3", Object: Object{Size: 100, Hash: "ghi"}}
	manifest.Files = append(manifest.Files, newFile)
	err = CheckManifestLimits(&manifest, &limits)
	if err == nil {
		t.Fatal()
	}
}
