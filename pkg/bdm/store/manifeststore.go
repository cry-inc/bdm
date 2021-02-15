package store

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"git.caputo.de/macaputo/bdm/pkg/bdm"
	"git.caputo.de/macaputo/bdm/pkg/bdm/util"
)

const manifestFileName = "manifest.json"

var manifestsMutex sync.RWMutex

// Call this method only if you have already locked the manifestsMutex exclusively!
func (s packageStore) addManifestLocked(manifest *bdm.Manifest) error {
	err := bdm.ValidatePublishedManifest(manifest)
	if err != nil {
		return fmt.Errorf("error validating published manifest: %w", err)
	}

	if !util.FolderExists(s.manifestsFolder) {
		return fmt.Errorf("manifest store folder does not exist")
	}

	packageFolder := path.Join(s.manifestsFolder, manifest.PackageName)
	versionFolder := path.Join(packageFolder, strconv.FormatUint(uint64(manifest.PackageVersion), 10))
	if util.FolderExists(versionFolder) {
		return fmt.Errorf("manifest with package name %s and version %d already exists",
			manifest.PackageName, manifest.PackageVersion)
	}

	err = os.MkdirAll(versionFolder, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating manifest folder: %w", err)
	}

	jsonData, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("error marshalling manifest to JSON: %w", err)
	}

	manifestPath := path.Join(versionFolder, manifestFileName)
	fileHandle, err := os.Create(manifestPath)
	if err != nil {
		return fmt.Errorf("error opening manifest file %s: %w", manifestPath, err)
	}
	defer fileHandle.Close()

	_, err = fileHandle.Write(jsonData)
	if err != nil {
		return fmt.Errorf("error writing manifest data: %w", err)
	}

	return nil
}

func (s packageStore) searchDuplicate(manifest *bdm.Manifest) error {
	existingVersions, err := s.GetVersions(manifest.PackageName)
	if err != nil {
		return fmt.Errorf("error getting versions for package %s: %w",
			manifest.PackageName, err)
	}
	for _, version := range existingVersions {
		existingManifest, err := s.GetManifest(manifest.PackageName, version)
		if err != nil {
			return fmt.Errorf("error getting manifest for package %s version %d: %w",
				manifest.PackageName, version, err)
		}
		if len(existingManifest.Files) != len(manifest.Files) {
			// Different number of files -> no duplicate
			continue
		}

		allFilesFoundAndIdentical := true
		for _, existingFile := range existingManifest.Files {
			found := false
			for _, file := range manifest.Files {
				if existingFile.Path == file.Path &&
					existingFile.Object.Hash == file.Object.Hash &&
					existingFile.Object.Size == file.Object.Size {
					found = true
					break
				}
			}
			if !found {
				allFilesFoundAndIdentical = false
				break
			}
		}

		if allFilesFoundAndIdentical {
			err := fmt.Errorf("found identical older version %d for package %s",
				version, manifest.PackageName)
			return DuplicatePackageError{err}
		}
	}

	return nil
}

func (s packageStore) AddManifest(manifest *bdm.Manifest) error {
	manifestsMutex.Lock()
	defer manifestsMutex.Unlock()

	return s.addManifestLocked(manifest)
}

func (s packageStore) PublishManifest(manifest *bdm.Manifest) error {
	err := bdm.ValidateUnpublishedManifest(manifest)
	if err != nil {
		return fmt.Errorf("error validating unpublished manifest: %w", err)
	}

	err = s.searchDuplicate(manifest)
	if err != nil {
		return fmt.Errorf("error searching for duplicate package: %w", err)
	}

	manifestsMutex.Lock()
	defer manifestsMutex.Unlock()

	var newVersion uint = 1
	existingVersions, err := s.GetVersions(manifest.PackageName)
	if err != nil {
		return fmt.Errorf("error getting existing versions for package %s: %w",
			manifest.PackageName, err)
	}
	for _, version := range existingVersions {
		if version >= newVersion {
			newVersion = version + 1
		}
	}
	manifest.PackageVersion = newVersion
	manifest.Published = time.Now().Unix()
	manifest.Hash = bdm.HashManifest(manifest)

	err = bdm.ValidatePublishedManifest(manifest)
	if err != nil {
		return fmt.Errorf("error validating published manifest: %w", err)
	}

	return s.addManifestLocked(manifest)
}

func (s packageStore) GetManifest(packageName string, version uint) (*bdm.Manifest, error) {
	if !util.FolderExists(s.manifestsFolder) {
		return nil, fmt.Errorf("manifest store folder does not exist")
	}

	manifestsMutex.RLock()
	defer manifestsMutex.RUnlock()

	packageFolder := path.Join(s.manifestsFolder, packageName)
	if !util.FolderExists(packageFolder) {
		return nil, fmt.Errorf("package %s does not exist", packageName)
	}

	versionFolder := path.Join(packageFolder, strconv.FormatUint(uint64(version), 10))
	if !util.FolderExists(versionFolder) {
		return nil, fmt.Errorf("package %s in version %d does not exist", packageName, version)
	}

	manifestPath := path.Join(versionFolder, manifestFileName)
	jsonData, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("error reading manifest file for package %s in version %d: %w",
			packageName, version, err)
	}

	var manifest bdm.Manifest
	err = json.Unmarshal(jsonData, &manifest)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling manifest JSON: %w", err)
	}

	return &manifest, nil
}

func (s packageStore) GetNames() ([]string, error) {
	if !util.FolderExists(s.manifestsFolder) {
		return nil, fmt.Errorf("manifest store folder does not exist")
	}

	items, err := ioutil.ReadDir(s.manifestsFolder)
	if err != nil {
		return nil, fmt.Errorf("error reading manifest store directory: %w", err)
	}

	names := make([]string, 0)
	for _, item := range items {
		if item.IsDir() {
			name := item.Name()
			names = append(names, name)
		}
	}

	return names, nil
}

func (s packageStore) GetVersions(packageName string) ([]uint, error) {
	if !util.FolderExists(s.manifestsFolder) {
		return nil, fmt.Errorf("manifest store folder does not exist")
	}

	packageFolder := path.Join(s.manifestsFolder, packageName)
	if !util.FolderExists(packageFolder) {
		return []uint{}, nil
	}

	files, err := ioutil.ReadDir(packageFolder)
	if err != nil {
		return nil, fmt.Errorf("error reading manifest store directory: %w", err)
	}

	versions := make([]uint, 0)
	for _, file := range files {
		if file.IsDir() {
			name := file.Name()
			i, err := strconv.Atoi(name)
			if err != nil {
				return nil, fmt.Errorf("error collecting version numbers from folder names: %w", err)
			}
			u := uint(i)
			versions = append(versions, u)
		}
	}

	return versions, nil
}
