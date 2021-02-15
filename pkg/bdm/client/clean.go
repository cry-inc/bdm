package client

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cry-inc/bdm/pkg/bdm"
)

func hasFile(relPath string, manifest *bdm.Manifest) bool {
	for _, file := range manifest.Files {
		if file.Path == relPath {
			return true
		}
	}
	return false
}

func hasFolder(relPath string, manifest *bdm.Manifest) bool {
	for _, file := range manifest.Files {
		if strings.Index(file.Path, relPath+"/") == 0 {
			return true
		}
	}
	return false
}

// CleanPackage will delete all non-package files from a packae folder
func CleanPackage(manifest *bdm.Manifest, packageFolder string) error {
	return filepath.Walk(packageFolder, func(filePath string, fileInfo os.FileInfo, err error) error {
		if os.IsNotExist(err) {
			// Ignore not existing files.
			// Files that are not there can also not contaminate a package :)
			// This happens when we delete whole directories inside this visitor.
			return nil
		}
		if err != nil {
			return fmt.Errorf("error while walking over folder %s: %w",
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
		relPath = filepath.ToSlash(relPath)
		if strings.Index(relPath, ".") == 0 || strings.Index(relPath, "..") != -1 {
			return nil
		}
		fileMode := fileInfo.Mode()
		remove := false
		if fileMode.IsDir() {
			if !hasFolder(relPath, manifest) {
				remove = true
			}
		} else if fileMode.IsRegular() {
			if !hasFile(relPath, manifest) {
				remove = true
			}
		} else {
			return fmt.Errorf("found unexpected item %s which is not a regular file or folder",
				relPath)
		}
		if remove {
			err = os.RemoveAll(filePath)
			if err != nil {
				return fmt.Errorf("error cleaning path %s: %w", filePath, err)
			}
		}
		return nil
	})
}
