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
	"github.com/cry-inc/bdm/pkg/bdm/util"
)

const apiKeyField = "bdm-api-key"

// UploadPackage publishes the specified folder as package to a remote server.
// This includes uploading of all files that doe not yet exists on the server.
func UploadPackage(name, inputFolder, serverURL, apiKey string) (*bdm.Manifest, error) {
	manifest, err := bdm.GenerateManifest(name, inputFolder)
	if err != nil {
		return nil, fmt.Errorf("error generating manifest for folder %s: %w",
			inputFolder, err)
	}

	err = bdm.ValidateUnpublishedManifest(manifest)
	if err != nil {
		return nil, fmt.Errorf("error validating generated manifest: %w", err)
	}

	err = checkRemoteManifestLimits(manifest, serverURL)
	if err != nil {
		return nil, fmt.Errorf("manifest failed to pass check against server limits: %w", err)
	}

	missingFiles, err := findFilesToUpload(manifest, serverURL)
	if err != nil {
		return nil, fmt.Errorf("error finding files to upload: %w", err)
	}

	if len(missingFiles) > 0 {
		err = uploadFiles(missingFiles, inputFolder, serverURL, apiKey)
		if err != nil {
			return nil, fmt.Errorf("error uploading files: %w", err)
		}
	}

	publishedManifest, err := publishManifest(manifest, serverURL, apiKey)
	if err != nil {
		return nil, fmt.Errorf("error publishing manifest: %w", err)
	}

	return publishedManifest, nil
}

func getObjects(files []bdm.File) []bdm.Object {
	objects := make([]bdm.Object, 0)
	for _, file := range files {
		objects = append(objects, file.Object)
	}
	return objects
}

func filterDuplicateFileObjects(files []bdm.File) []bdm.File {
	objectMap := make(map[string]bool)
	filtered := make([]bdm.File, 0)
	for _, file := range files {
		if _, found := objectMap[file.Object.Hash]; !found {
			objectMap[file.Object.Hash] = true
			filtered = append(filtered, file)
		}
	}
	return filtered
}

func findFilesToUpload(manifest *bdm.Manifest, serverURL string) ([]bdm.File, error) {
	objects := getObjects(manifest.Files)

	r, w := io.Pipe()
	go func() {
		defer w.Close()
		compressedWriter, err := util.CreateCompressingWriter(w)
		if err != nil {
			panic(fmt.Errorf("error creating compressing writer: %w", err))
		}
		defer compressedWriter.Close()
		err = bdm.WriteObjectsToStream(objects, compressedWriter)
		if err != nil {
			panic(fmt.Errorf("error writing objects to stream: %w", err))
		}
	}()

	url := serverURL + "/objects/check"
	req, err := http.NewRequest("POST", url, r)
	if err != nil {
		return nil, fmt.Errorf("error creating POST request for URL %s: %w", url, err)
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending POST request to URL %s: %w", url, err)
	}
	defer res.Body.Close()

	resData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading POST response body: %w", err)
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("error checking objects: server returned status code %d: %s",
			res.StatusCode, resData)
	}

	var foundObjects []bdm.Object
	err = json.Unmarshal(resData, &foundObjects)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling objects from JSON: %w", err)
	}

	filesToUpload := make([]bdm.File, 0)
	for _, file := range manifest.Files {
		found := false
		for _, foundObject := range foundObjects {
			if file.Object.Hash == foundObject.Hash {
				found = true
				break
			}
		}
		if !found {
			filesToUpload = append(filesToUpload, file)
		}
	}

	// Remove files with duplicate file objects!
	filesToUpload = filterDuplicateFileObjects(filesToUpload)

	return filesToUpload, nil
}

func uploadFiles(files []bdm.File, inputFolder, serverURL, apiKey string) error {
	r, w := io.Pipe()
	go func() {
		defer w.Close()
		compressedWriter, err := util.CreateCompressingWriter(w)
		if err != nil {
			panic(fmt.Errorf("error creating compressing writer: %w", err))
		}
		defer compressedWriter.Close()

		objects := getObjects(files)
		err = bdm.WriteObjectsToStream(objects, compressedWriter)
		if err != nil {
			panic(fmt.Errorf("error writing objects to stream: %w", err))
		}

		for _, file := range files {
			fullPath := filepath.Join(inputFolder, file.Path)
			fileHandle, err := os.Open(fullPath)
			if err != nil {
				panic(fmt.Errorf("error opening file %s: %w", fullPath, err))
			}
			defer fileHandle.Close()

			copied, err := io.Copy(compressedWriter, fileHandle)
			if err != nil {
				panic(fmt.Errorf("error copying file %s: %w", fullPath, err))
			}
			if copied != file.Object.Size {
				panic(fmt.Errorf("error reading file %s: expected %d but found %d bytes",
					fullPath, file.Object.Size, copied))
			}
		}
	}()

	url := serverURL + "/objects/upload"
	req, err := http.NewRequest("POST", url, r)
	if err != nil {
		return fmt.Errorf("error creating POST request for URL %s: %w", url, err)
	}
	req.Header.Add(apiKeyField, apiKey)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending POST request to URL %s: %w", url, err)
	}
	defer res.Body.Close()

	resData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error reading POST response from URL %s: %w", url, err)
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("error posting objects: server returned status code %d: %s",
			res.StatusCode, resData)
	}

	var uploadedObjects []bdm.Object
	err = json.Unmarshal(resData, &uploadedObjects)
	if err != nil {
		return fmt.Errorf("error unmarshalling JSON objects: %w", err)
	}

	// More checking required?
	if len(uploadedObjects) != len(files) {
		return fmt.Errorf("error uploading files: uploaded %d objects but expected to upload %d",
			len(uploadedObjects), len(files))
	}

	return nil
}

func publishManifest(manifest *bdm.Manifest, serverURL, apiKey string) (*bdm.Manifest, error) {
	jsonData, err := json.Marshal(manifest)
	if err != nil {
		return nil, fmt.Errorf("error marshalling manifest to JSON: %w", err)
	}

	r, w := io.Pipe()
	go func() {
		defer w.Close()
		written, err := w.Write(jsonData)
		if err != nil {
			panic(fmt.Errorf("failed to write JSON to stream: %w", err))
		}
		if written != len(jsonData) {
			panic(fmt.Errorf("failed to write all JSON data to stream"))
		}
	}()

	url := serverURL + "/manifests"
	req, err := http.NewRequest("POST", url, r)
	if err != nil {
		return nil, fmt.Errorf("error creating POST request for URL %s: %w", url, err)
	}
	req.Header.Add(apiKeyField, apiKey)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending POST request to URL %s: %w", url, err)
	}
	defer res.Body.Close()

	resData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading POST response body: %w", err)
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("error posting manifest: server returned status code %d: %s",
			res.StatusCode, resData)
	}

	var publishedManifest bdm.Manifest
	err = json.Unmarshal(resData, &publishedManifest)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling manifest JSON: %w", err)
	}

	err = bdm.ValidatePublishedManifest(&publishedManifest)
	if err != nil {
		return nil, fmt.Errorf("error validating published manifest: %w", err)
	}

	if manifest.PackageName != publishedManifest.PackageName ||
		len(manifest.Files) != len(publishedManifest.Files) {
		return nil, fmt.Errorf("detected manifest mismatch for package name or number of files")
	}

	return &publishedManifest, nil
}

func checkRemoteManifestLimits(manifest *bdm.Manifest, serverURL string) error {
	url := serverURL + "/limits"
	res, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error getting limits from remote server at %s: %w", url, err)
	}

	defer res.Body.Close()
	resData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error reading limits response body: %w", err)
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("error getting server limits: server returned status code %d: %s",
			res.StatusCode, resData)
	}

	var limits bdm.ManifestLimits
	err = json.Unmarshal(resData, &limits)
	if err != nil {
		return fmt.Errorf("error unmarshalling limits JSON: %w", err)
	}

	return bdm.CheckManifestLimits(manifest, &limits)
}
