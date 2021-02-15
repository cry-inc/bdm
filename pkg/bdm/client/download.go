package client

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cry-inc/bdm/pkg/bdm"
	"github.com/cry-inc/bdm/pkg/bdm/store"
	"github.com/cry-inc/bdm/pkg/bdm/util"
)

// DownloadPackage downloads a package from a remote server to a local folder
func DownloadPackage(outputFolder, serverURL, name string, version uint, clean bool) error {
	manifest, err := DownloadManifest(serverURL, name, version)
	if err != nil {
		return fmt.Errorf("error downloading manifest: %w", err)
	}

	err = DownloadFiles(serverURL, manifest, outputFolder)
	if err != nil {
		return fmt.Errorf("error downloading files: %w", err)
	}

	if clean {
		err = CleanPackage(manifest, outputFolder)
		if err != nil {
			return fmt.Errorf("error cleaning package output folder: %w", err)
		}
	}

	return nil
}

// DownloadCachedPackage is like DownloadPackage with an additional local cache
func DownloadCachedPackage(outputFolder, cacheFolder, serverURL, name string, version uint, clean bool) error {
	manifest, err := DownloadCachedManifest(cacheFolder, serverURL, name, version)
	if err != nil {
		return fmt.Errorf("error downloading cached manifest: %w", err)
	}

	err = DownloadCachedFiles(cacheFolder, serverURL, manifest, outputFolder)
	if err != nil {
		return fmt.Errorf("error downloading cached files: %w", err)
	}

	if clean {
		err = CleanPackage(manifest, outputFolder)
		if err != nil {
			return fmt.Errorf("error cleaning package output folder: %w", err)
		}
	}

	return nil
}

// DownloadManifest fetches the specified package manifest from a server
func DownloadManifest(serverURL, name string, version uint) (*bdm.Manifest, error) {
	url := fmt.Sprintf("%s/manifests/%s/%d", serverURL, name, version)
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error getting URL %s: %w", url, err)
	}
	defer res.Body.Close()

	resData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading manifest body: %w", err)
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("error getting URL %s: server returned status code %d: %s",
			url, res.StatusCode, resData)
	}

	var manifest bdm.Manifest
	err = json.Unmarshal(resData, &manifest)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling manifest JSON: %w", err)
	}

	err = bdm.ValidatePublishedManifest(&manifest)
	if err != nil {
		return nil, fmt.Errorf("error validating received manifest: %w", err)
	}

	return &manifest, nil
}

// DownloadCachedManifest is DownloadManifest with an additional integrated cache
func DownloadCachedManifest(cacheFolder, serverURL, name string, version uint) (*bdm.Manifest, error) {
	store, err := store.New(cacheFolder)
	if err != nil {
		return nil, fmt.Errorf("failed to open cache store at location %s: %w", cacheFolder, err)
	}

	manifest, err := store.GetManifest(name, version)
	if err == nil {
		// Early out, manifest found!
		return manifest, nil
	}

	manifest, err = DownloadManifest(serverURL, name, version)
	if err != nil {
		return nil, fmt.Errorf("error downloading manifest: %w", err)
	}

	err = store.AddManifest(manifest)
	if err != nil {
		return nil, fmt.Errorf("error adding manifest to cache store: %w", err)
	}

	return manifest, nil
}

func getMissingFiles(manifest *bdm.Manifest, outputFolder string) []bdm.File {
	missingFiles := make([]bdm.File, 0)
	for _, file := range manifest.Files {
		fullPath := filepath.Join(outputFolder, file.Path)
		fileInfo, err := os.Stat(fullPath)
		if err == nil && !fileInfo.IsDir() && fileInfo.Mode().IsRegular() {
			fileSize := fileInfo.Size()
			if fileSize == file.Object.Size {
				hash, err := util.HashFile(fullPath)
				if err == nil && hash == file.Object.Hash {
					// File exists with the correct size and hash,
					// no changes needed and we can skip this one :)
					continue
				}
			}
		}
		missingFiles = append(missingFiles, file)
	}
	return missingFiles
}

func filterFilesByObject(files []bdm.File, objectHash string) []bdm.File {
	filtered := make([]bdm.File, 0)
	for _, file := range files {
		if file.Object.Hash == objectHash {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

func writeObjectToFiles(reader io.Reader, files []bdm.File, outputFolder string) error {
	if len(files) < 1 {
		return fmt.Errorf("found invalid number of files: %d", len(files))
	}

	// Create folders for all files
	for i, file := range files {
		fullPath := filepath.Join(outputFolder, file.Path)
		folder := filepath.Dir(fullPath)
		if !util.FolderExists(folder) {
			err := os.MkdirAll(folder, os.ModePerm)
			if err != nil {
				return fmt.Errorf("error creating directory %s: %w", folder, err)
			}
		}
		if i == 0 {
			// First file is special, since its streamed over the network
			fileHandle, err := os.Create(fullPath)
			if err != nil {
				return fmt.Errorf("error creating file %s: %w", fullPath, err)
			}
			defer fileHandle.Close()

			hasher := util.CreateHasher()
			writer := io.MultiWriter(fileHandle, hasher)

			written, err := io.CopyN(writer, reader, file.Object.Size)
			if err != nil {
				return fmt.Errorf("error writing object data for file %s: %w", file.Path, err)
			}
			if written != file.Object.Size {
				return fmt.Errorf("error writing object data for file %s: received %d but expected %d bytes",
					file.Path, written, file.Object.Size)
			}

			hash := util.GetHashString(hasher)
			if hash != file.Object.Hash {
				return fmt.Errorf("error writing object data for file %s: found hash %s but expected %s",
					file.Path, hash, file.Object.Hash)
			}
		} else {
			// All the other files are copied locallay since they have the same content!
			fullSourcePath := filepath.Join(outputFolder, files[0].Path)
			err := copyFile(fullSourcePath, fullPath, file.Object.Hash)
			if err != nil {
				return fmt.Errorf("error copying file %s to %s: %w", fullSourcePath, fullPath, err)
			}
		}
	}

	return nil
}

func copyFile(source, target, hash string) error {
	readHandle, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("Unable to open source file %s: %w", source, err)
	}
	defer readHandle.Close()

	writeHandle, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("Unable to open target file %s: %w", target, err)
	}
	defer writeHandle.Close()

	hasher := util.CreateHasher()
	writer := io.MultiWriter(writeHandle, hasher)

	_, err = io.Copy(writer, readHandle)
	if err != nil {
		return fmt.Errorf("Unable to copy file data: %w", err)
	}
	if util.GetHashString(hasher) != hash {
		return fmt.Errorf("Failed to verify file content: Hash mismatch")
	}

	return nil
}

// DownloadFiles downloads all all files from a manifest to a output folder.
// It skips all files that already exists in the folder with the correct size and hash.
func DownloadFiles(serverURL string, manifest *bdm.Manifest, outputFolder string) error {
	missingFiles := getMissingFiles(manifest, outputFolder)

	if len(missingFiles) == 0 {
		// Early out, all files already there!
		return nil
	}

	// Filter duplicate file objects to download them only once!
	filteredFiles := filterDuplicateFileObjects(missingFiles)
	requiredObjects := getObjects(filteredFiles)

	r, w := io.Pipe()
	go func() {
		defer w.Close()
		compressedWriter, err := util.CreateCompressingWriter(w)
		if err != nil {
			panic(fmt.Errorf("failed to create compressing writer: %w", err))
		}
		defer compressedWriter.Close()
		err = bdm.WriteObjectsToStream(requiredObjects, compressedWriter)
		if err != nil {
			panic(fmt.Errorf("failed write objects to stream: %w", err))
		}
	}()

	url := serverURL + "/objects/download"
	req, err := http.NewRequest("POST", url, r)
	if err != nil {
		return fmt.Errorf("error creating POST request for URL %s: %w", url, err)
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error posting URL %s: %w", url, err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("error posting URL %s: server returned status code %d",
			url, res.StatusCode)
	}

	reader, err := util.CreateDecompressingReader(res.Body)
	if err != nil {
		return fmt.Errorf("error creating decompressing reader: %w", err)
	}
	defer reader.Close()

	incomingObjects, err := bdm.ReadObjectsFromStream(reader)
	if err != nil {
		return fmt.Errorf("error reading object stream from server: %w", err)
	}

	if len(incomingObjects) != len(requiredObjects) {
		return fmt.Errorf("error reading objects from server: received %d objects but expected %d",
			len(incomingObjects), len(requiredObjects))
	}

	for _, object := range incomingObjects {
		// Get list of missing files that needs this object as content
		files := filterFilesByObject(missingFiles, object.Hash)

		// Stream objects files to the output folder and verify them
		err := writeObjectToFiles(reader, files, outputFolder)
		if err != nil {
			return fmt.Errorf("error writing object to files in output folder: %w", err)
		}
	}

	return nil
}

// DownloadCachedFiles is DownloadFiles with an integrated cache
func DownloadCachedFiles(cacheFolder, serverURL string, manifest *bdm.Manifest, outputFolder string) error {
	cache, err := store.New(cacheFolder)
	if err != nil {
		return fmt.Errorf("failed to open cache store at location %s: %w", cacheFolder, err)
	}

	err = restoreFilesFromCache(cache, manifest, outputFolder)
	if err != nil {
		return fmt.Errorf("error restoring files from cache: %w", err)
	}

	err = DownloadFiles(serverURL, manifest, outputFolder)
	if err != nil {
		return fmt.Errorf("error downloading files: %w", err)
	}

	err = addFilesToCache(cache, manifest, outputFolder)
	if err != nil {
		return fmt.Errorf("error adding files to cache: %w", err)
	}

	return nil
}

func restoreFilesFromCache(cache store.Store, manifest *bdm.Manifest, outputFolder string) error {
	missingFiles := getMissingFiles(manifest, outputFolder)
	for _, file := range missingFiles {
		object, err := cache.GetObject(file.Object.Hash)
		if err == nil && object.Size == file.Object.Size && object.Hash == file.Object.Hash {
			fullPath := filepath.Join(outputFolder, file.Path)
			folder := filepath.Dir(fullPath)
			if !util.FolderExists(folder) {
				err := os.MkdirAll(folder, os.ModePerm)
				if err != nil {
					return fmt.Errorf("error creating directory %s: %w", folder, err)
				}
			}
			fileHandle, err := os.Create(fullPath)
			if err != nil {
				return fmt.Errorf("error creating file %s: %w", fullPath, err)
			}
			defer fileHandle.Close()

			reader, err := cache.ReadObject(object.Hash)
			if err != nil {
				return fmt.Errorf("error reading object %s from cache: %w", object.Hash, err)
			}
			defer reader.Close()

			hasher := util.CreateHasher()
			writer := io.MultiWriter(fileHandle, hasher)

			written, err := io.Copy(writer, reader)
			if err != nil {
				return fmt.Errorf("error writing object data: %w", err)
			}

			if written != object.Size {
				return fmt.Errorf("error writing object data for file %s: received %d but expected %d bytes",
					file.Path, written, file.Object.Size)
			}

			hash := util.GetHashString(hasher)
			if hash != file.Object.Hash {
				return fmt.Errorf("error writing object data for file %s: found hash %s but expected %s",
					file.Path, hash, file.Object.Hash)
			}
		}
	}
	return nil
}

func addFilesToCache(cache store.Store, manifest *bdm.Manifest, outputFolder string) error {
	for _, file := range manifest.Files {
		fullPath := filepath.Join(outputFolder, file.Path)

		fileHandle, err := os.Open(fullPath)
		if err != nil {
			return fmt.Errorf("error opening file %s: %w", fullPath, err)
		}
		defer fileHandle.Close()

		object, err := cache.AddObject(fileHandle)
		if err != nil {
			return fmt.Errorf("error adding file %s to cache: %w", file.Path, err)
		}

		if object.Size != file.Object.Size {
			return fmt.Errorf("error adding file %s to cache: %w: expected %d but found %d bytes",
				file.Path, err, file.Object.Size, object.Size)
		}

		if object.Hash != file.Object.Hash {
			return fmt.Errorf("error adding file %s to cache: %w: expected hash %s but found %s",
				file.Path, err, file.Object.Hash, object.Hash)
		}
	}

	return nil
}
