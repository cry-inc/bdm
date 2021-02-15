package bdm

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cry-inc/bdm/pkg/bdm/util"
)

// Object represents the content of a package file
type Object struct {
	Size int64
	Hash string
}

// File represents a file that is part of a package
type File struct {
	Path   string
	Object Object
}

// A Manifest is a complete description of a package
type Manifest struct {
	ManifestVersion uint
	PackageName     string
	PackageVersion  uint
	Published       int64
	Hash            string
	Files           []File
}

// GenerateManifest creates an unpublished manifest for an input folder using the given name
func GenerateManifest(packageName, inputFolder string) (*Manifest, error) {
	files := make([]File, 0)
	err := filepath.Walk(inputFolder, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error getting file info: %w", err)
		}

		fileMode := fileInfo.Mode()
		if !fileMode.IsDir() && fileMode.IsRegular() {
			absInput, err := filepath.Abs(inputFolder)
			if err != nil {
				return fmt.Errorf("error getting absolute input path for %s: %w",
					inputFolder, err)
			}

			absFile, err := filepath.Abs(filePath)
			if err != nil {
				return fmt.Errorf("error getting absolute file path for %s: %w",
					filePath, err)
			}

			packageFilePath, err := filepath.Rel(absInput, absFile)
			if err != nil {
				return fmt.Errorf("error getting relative file path between %s and %s: %w",
					absInput, absFile, err)
			}

			packageFilePath = filepath.ToSlash(packageFilePath)
			hash, err := util.HashFile(filePath)
			if err != nil {
				return fmt.Errorf("error hashing file %s: %w", filePath, err)
			}

			packageFile := File{
				Path: packageFilePath,
				Object: Object{
					Size: fileInfo.Size(),
					Hash: hash,
				},
			}
			files = append(files, packageFile)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking over input folder %s: %w",
			inputFolder, err)
	}

	manifest := Manifest{
		ManifestVersion: 1,
		PackageName:     packageName,
		Files:           files,
	}

	manifest.Hash = HashManifest(&manifest)

	return &manifest, nil
}

// ValidatePackageName will return true for valid package names
func ValidatePackageName(name string) bool {
	validName, _ := regexp.MatchString(`^[a-z0-9_-]+$`, name)
	return validName
}

func validateBasicManifest(manifest *Manifest) error {
	if manifest.ManifestVersion != 1 {
		return fmt.Errorf("invalid manifest version")
	}
	if !ValidatePackageName(manifest.PackageName) {
		return fmt.Errorf("invalid package name")
	}
	if len(manifest.Files) <= 0 {
		return fmt.Errorf("manifest contains no files")
	}

	paths := make(map[string]bool)
	for _, file := range manifest.Files {
		if len(file.Path) == 0 {
			return fmt.Errorf("found empty file path")
		}
		if strings.Index(file.Path, "..") != -1 {
			return fmt.Errorf("invalid file path %s", file.Path)
		}
		if file.Object.Size < 0 {
			return fmt.Errorf("invalid object size %d for file %s", file.Object.Size, file.Path)
		}
		validHash, _ := regexp.MatchString(`^[a-f0-9_-]+$`, file.Object.Hash)
		if !validHash {
			return fmt.Errorf("invalid object hash %s", file.Object.Hash)
		}
		// duplicates are checked case-insenstitive to avoid Windows issues
		lowerCasePath := strings.ToLower(file.Path)
		if paths[lowerCasePath] {
			return fmt.Errorf("duplicate file path %s", file.Path)
		}
		paths[lowerCasePath] = true
	}

	hash := HashManifest(manifest)
	if hash != manifest.Hash {
		return fmt.Errorf("invalid manifest hash")
	}

	return nil
}

// ValidateUnpublishedManifest can validate unpublished manifests and will return nil when no issues where found
func ValidateUnpublishedManifest(manifest *Manifest) error {
	err := validateBasicManifest(manifest)
	if err != nil {
		return fmt.Errorf("error validating basic manifest data: %w", err)
	}
	if manifest.PackageVersion != 0 {
		return fmt.Errorf("package version is not zero")
	}
	if manifest.Published != 0 {
		return fmt.Errorf("published date is not zero")
	}
	return nil
}

// ValidatePublishedManifest can validate published manifests and will return nil when no issues where found
func ValidatePublishedManifest(manifest *Manifest) error {
	err := validateBasicManifest(manifest)
	if err != nil {
		return err
	}
	if manifest.PackageVersion <= 0 {
		return fmt.Errorf("invalid package version")
	}
	if manifest.Published <= 0 {
		return fmt.Errorf("invalid published date")
	}
	return nil
}

// HashManifest calculates the verification hash for a manifest
func HashManifest(manifest *Manifest) string {
	hasher := util.CreateHasher()
	addString := func(s string) {
		hasher.Write([]byte(s))
	}

	addString(fmt.Sprint(manifest.ManifestVersion))
	addString(manifest.PackageName)
	addString(fmt.Sprint(manifest.ManifestVersion))
	addString(fmt.Sprint(manifest.Published))

	for _, file := range manifest.Files {
		addString(file.Path)
		addString(file.Object.Hash)
		addString(fmt.Sprint(file.Object.Size))
	}

	return util.GetHashString(hasher)
}
