package store

import (
	"fmt"
	"io"
	"os"
	"path"
	"sync"

	"github.com/cry-inc/bdm/pkg/bdm"
	"github.com/cry-inc/bdm/pkg/bdm/util"
)

// DuplicatePackageError is returned when trying to publish a package
// with content already existing in an earlier version of the same package.
type DuplicatePackageError struct{ error }

// Store represents a persitent store for package data
type Store interface {
	PublishManifest(manifest *bdm.Manifest) error
	AddManifest(manifest *bdm.Manifest) error
	GetNames() ([]string, error)
	GetVersions(packageName string) ([]uint, error)
	GetManifest(packageName string, version uint) (*bdm.Manifest, error)

	GetObject(hash string) (*bdm.Object, error)
	AddObject(reader io.Reader) (*bdm.Object, error)
	ReadObject(hash string) (io.ReadCloser, error)
	GetObjects() ([]*bdm.Object, error)
}

type packageStore struct {
	manifestsFolder string
	objectsFolder   string
	objectsMutex    sync.Mutex
	manifestsMutex  sync.RWMutex
}

const manifestsSubFolder = "manifests"
const objectsSubFolder = "objects"

// New creates a new persistent filesystem-based package store
func New(storeFolder string) (Store, error) {
	if !util.FolderExists(storeFolder) {
		err := os.MkdirAll(storeFolder, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("error creating folder %s for new store: %w",
				storeFolder, err)
		}
	}

	var store packageStore
	store.manifestsFolder = path.Join(storeFolder, manifestsSubFolder)
	if !util.FolderExists(store.manifestsFolder) {
		err := os.Mkdir(store.manifestsFolder, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("error creating manifests folder %s for new store: %w",
				store.manifestsFolder, err)
		}
	}
	store.objectsFolder = path.Join(storeFolder, objectsSubFolder)
	if !util.FolderExists(store.objectsFolder) {
		err := os.Mkdir(store.objectsFolder, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("error creating objects folder %s for new store: %w",
				store.objectsFolder, err)
		}
	}

	return &store, nil
}

// AllObjectsExist can verify that all objects from the manifest exist in the store
func AllObjectsExist(manifest *bdm.Manifest, store Store) bool {
	for _, f := range manifest.Files {
		_, err := store.GetObject(f.Object.Hash)
		if err != nil {
			return false
		}
	}
	return true
}

func getAllManifests(store Store) ([]*bdm.Manifest, error) {
	manifests := make([]*bdm.Manifest, 0)
	names, err := store.GetNames()
	if err != nil {
		return nil, fmt.Errorf("error getting manifest names: %w", err)
	}
	for _, name := range names {
		versions, err := store.GetVersions(name)
		if err != nil {
			return nil, fmt.Errorf("error getting package versions for %s: %w", name, err)
		}
		for _, version := range versions {
			manifest, err := store.GetManifest(name, version)
			if err != nil {
				return nil, fmt.Errorf("error getting manifest %s version %d: %w",
					name, version, err)
			}
			manifests = append(manifests, manifest)
		}
	}
	return manifests, nil
}

func checkObject(store Store, object *bdm.Object) error {
	reader, err := store.ReadObject(object.Hash)
	if err != nil {
		return fmt.Errorf("error reading object %s: %w", object.Hash, err)
	}
	defer reader.Close()

	hasher := util.CreateHasher()
	read, err := io.Copy(hasher, reader)
	if err != nil {
		return fmt.Errorf("error hashing object %s: %w", object.Hash, err)
	}
	if read != object.Size {
		return fmt.Errorf("found size mismatch for object %s: expected %d but read %d bytes",
			object.Hash, object.Size, read)
	}

	hash := util.GetHashString(hasher)
	if hash != object.Hash {
		return fmt.Errorf("found hash mismatch for object %s: expected %s but found %s",
			object.Hash, object.Hash, hash)
	}

	return nil
}

// ValidateStore validates the whole package store.
// It will validate all manifests,
// check that all the objects of all manifests exist and
// check that all objects are valid and produce the correct hash.
// It also returns a map with some simple store statistics.
func ValidateStore(store Store) (map[string]int64, error) {
	manifests, err := getAllManifests(store)
	if err != nil {
		return nil, fmt.Errorf("error listing all manifests: %w", err)
	}

	// Validate all manifests
	for _, manifest := range manifests {
		err = bdm.ValidatePublishedManifest(manifest)
		if err != nil {
			return nil, fmt.Errorf("error validating published manifest %s version %d: %w",
				manifest.PackageName, manifest.PackageVersion, err)
		}
	}

	// Get all object metadata
	objects, err := store.GetObjects()
	if err != nil {
		return nil, fmt.Errorf("error getting objects list from store: %w", err)
	}

	// Build map[hash]object for fast lookup
	// and use the loop to also sum up overall size
	objectsMap := make(map[string]*bdm.Object)
	var objectsSize int64 = 0
	for _, object := range objects {
		objectsMap[object.Hash] = object
		objectsSize += object.Size
	}

	// Check if all objects for all manifests exist
	for _, manifest := range manifests {
		for _, file := range manifest.Files {
			if _, ok := objectsMap[file.Object.Hash]; !ok {
				return nil, fmt.Errorf("unable to find object %s from package %s version %d",
					file.Object.Hash, manifest.PackageName, manifest.PackageVersion)
			}
		}
	}

	// Check all object data for consistency
	for _, object := range objects {
		err = checkObject(store, object)
		if err != nil {
			return nil, fmt.Errorf("found problem while checking object: %w", err)
		}
	}

	stats := make(map[string]int64)
	stats["packages"] = int64(len(manifests))
	stats["objects"] = int64(len(objects))
	stats["size"] = objectsSize

	return stats, nil
}
