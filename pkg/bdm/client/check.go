package client

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cry-inc/bdm/pkg/bdm"
	"github.com/cry-inc/bdm/pkg/bdm/util"
)

// CheckPackage will check an existing loack package folder against the manifest
func CheckPackage(packageFolder, serverURL, apiToken, name string, version uint, clean bool) error {
	manifest, err := DownloadManifest(serverURL, apiToken, name, version)
	if err != nil {
		return fmt.Errorf("error downloading manifest: %w", err)
	}

	return CheckFiles(manifest, packageFolder, clean)
}

// CheckCachedPackage is like CheckPackage but with an additional local cache
func CheckCachedPackage(packageFolder, cacheFolder, serverURL, apiToken, name string, version uint, clean bool) error {
	manifest, err := DownloadCachedManifest(cacheFolder, serverURL, apiToken, name, version)
	if err != nil {
		return fmt.Errorf("error downloading cached manifest: %w", err)
	}

	return CheckFiles(manifest, packageFolder, clean)
}

func checkFile(file bdm.File, packageFolder string) error {
	fullPath := filepath.Join(packageFolder, file.Path)
	if !util.FileExists(fullPath) {
		return fmt.Errorf("cannot find file %s", file.Path)
	}
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		return fmt.Errorf("error reading stats for file %s: %w",
			fullPath, err)
	}
	if fileInfo.Size() != file.Object.Size {
		return fmt.Errorf("file %s has the wrong size: expected %d and found %d bytes",
			file.Path, file.Object.Size, fileInfo.Size())
	}
	hash, err := util.HashFile(fullPath)
	if err != nil {
		return fmt.Errorf("error hashing file %s: %w",
			fullPath, err)
	}
	if hash != file.Object.Hash {
		return fmt.Errorf("file %s produced the wrong hash: expected %s and found %s",
			file.Path, file.Object.Hash, hash)
	}

	return nil
}

// CheckFiles compare a folder against a manifest and complain about missing or wrong files.
// It will also complain about non-package files if clean is set to true.
func CheckFiles(manifest *bdm.Manifest, packageFolder string, clean bool) error {
	for _, file := range manifest.Files {
		err := checkFile(file, packageFolder)
		if err != nil {
			return fmt.Errorf("found problem while check file: %w", err)
		}
	}

	if !clean {
		// No clean checks are requested -> early out
		return nil
	}

	return filepath.Walk(packageFolder, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking files in folder %s: %w",
				packageFolder, err)
		}
		absFolder, err := filepath.Abs(packageFolder)
		if err != nil {
			return fmt.Errorf("error getting absolute path for package folder %s: %w",
				packageFolder, err)
		}
		absFile, err := filepath.Abs(filePath)
		if err != nil {
			return fmt.Errorf("error getting absolute path for file %s: %w",
				filePath, err)
		}
		relPath, err := filepath.Rel(absFolder, absFile)
		if err != nil {
			return fmt.Errorf("error getting relative path between folder %s and file %s: %w",
				absFolder, absFile, err)
		}

		// Normalize (Windows) paths
		relPath = filepath.ToSlash(relPath)
		if strings.Index(relPath, ".") == 0 || strings.Contains(relPath, "..") {
			return nil
		}

		fileMode := fileInfo.Mode()
		if fileMode.IsDir() {
			if !hasFolder(relPath, manifest) {
				return fmt.Errorf("found non-package folder %s", relPath)
			}
		} else if fileMode.IsRegular() {
			if !hasFile(relPath, manifest) {
				return fmt.Errorf("found non-package file %s", relPath)
			}
		} else {
			return fmt.Errorf("found unexpected item %s which is not a regular file or folder",
				relPath)
		}
		return nil
	})
}
