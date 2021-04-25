package main

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cry-inc/bdm/pkg/bdm"
	"github.com/cry-inc/bdm/pkg/bdm/client"
	"github.com/cry-inc/bdm/pkg/bdm/server"
	"github.com/cry-inc/bdm/pkg/bdm/store"
	"github.com/cry-inc/bdm/pkg/bdm/util"
)

const token = "1234abcd"
const storeFolder = "test/store"
const outputFolder = "test/output"
const cacheFolder = "test/cache"
const serverURL = "http://127.0.0.1:2323"
const packageNameSmall = "foo"
const packageFolderSmall = "test/example"
const packageNameBig = "bar"
const packageFolderBig = "test/big"
const unzipFolder = "test/unzipped"

func TestServerClient(t *testing.T) {
	// Prepare Cleanup
	defer os.RemoveAll(storeFolder)
	defer os.RemoveAll(outputFolder)
	defer os.RemoveAll(cacheFolder)

	// Creation and cleanup of server store
	server, stopped := startTestingServer(t)

	// Publish a small test package
	publishSmallTestPackage(t)

	// Download a small test package
	err := client.DownloadPackage(outputFolder, serverURL, packageNameSmall, 1, false)
	if err != nil {
		t.Fatal(err)
	}

	// Check package content
	err = client.CheckPackage(outputFolder, serverURL, packageNameSmall, 1, true)
	if err != nil {
		t.Fatal(err)
	}

	// Create dirty file and download again with clean mode enabled
	const dirtyFile = outputFolder + "/dirty.dat"
	ioutil.WriteFile(dirtyFile, []byte{0, 1, 2}, os.ModePerm)
	err = client.DownloadPackage(outputFolder, serverURL, packageNameSmall, 1, true)
	if err != nil {
		t.Fatal(err)
	}
	if util.FileExists(dirtyFile) {
		t.Fatal()
	}

	// Download again with caching enabled
	os.RemoveAll(outputFolder)
	err = client.DownloadCachedPackage(outputFolder, cacheFolder, serverURL, packageNameSmall, 1, false)
	if err != nil {
		t.Fatal(err)
	}

	// Stop server
	stopTestingServer(server, stopped)

	// Try to restore package from cache only
	os.RemoveAll(outputFolder)
	err = client.DownloadCachedPackage(outputFolder, cacheFolder, serverURL, packageNameSmall, 1, true)
	if err != nil {
		t.Fatal(err)
	}

	// Check using the cache only
	err = client.CheckCachedPackage(outputFolder, cacheFolder, serverURL, packageNameSmall, 1, true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestServerJsonHandlers(t *testing.T) {
	// Prepare Cleanup
	defer os.RemoveAll(storeFolder)

	// Creation and cleanup of server store
	server, stopped := startTestingServer(t)
	defer stopTestingServer(server, stopped)

	// Check empty manifest list on fresh server
	getAndCompareString(t, "/manifests", "application/json", "[]")

	// Publish a small test package
	publishSmallTestPackage(t)

	// Look for fresh test package in the list
	getAndCompareString(t, "/manifests", "application/json", "[{\"Name\":\"foo\"}]")

	// List versions of the fresh test package
	getAndCompareString(t, "/manifests/foo", "application/json", "[{\"Version\":1}]")
}

func TestServerFileHandler(t *testing.T) {
	// Prepare Cleanup
	defer os.RemoveAll(storeFolder)

	// Creation and cleanup of server store
	server, stopped := startTestingServer(t)
	defer stopTestingServer(server, stopped)

	// Publish test package
	publishSmallTestPackage(t)

	// Small binary file
	expectedData := string([]byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0})
	urlPath := "/files/foo/1/213151e5833fecb107899dfd0c8baca0fb671d4017fbd9361c8007b7b93681a6/data.bin"
	getAndCompareString(t, urlPath, "application/octet-stream", expectedData)

	// Empty file
	urlPath = "/files/foo/1/af1349b9f5f9a1a6a0404dea36dcc9499bcb25c9adc112b7cc9a93cae41f3262/empty.txt"
	getAndCompareString(t, urlPath, "application/octet-stream", "")

	// Wrong file name
	urlPath = "/files/foo/1/213151e5833fecb107899dfd0c8baca0fb671d4017fbd9361c8007b7b93681a6/wrong.name"
	httpGetStatusCode(t, urlPath, 404)

	// Wrong package name
	urlPath = "/files/blaaaaa/1/213151e5833fecb107899dfd0c8baca0fb671d4017fbd9361c8007b7b93681a6/data.bin"
	httpGetStatusCode(t, urlPath, 404)

	// Invalid package name
	urlPath = "/files/bla+()aaaa/1/213151e5833fecb107899dfd0c8baca0fb671d4017fbd9361c8007b7b93681a6/data.bin"
	httpGetStatusCode(t, urlPath, 400)

	// Wrong package version
	urlPath = "/files/foo/666/213151e5833fecb107899dfd0c8baca0fb671d4017fbd9361c8007b7b93681a6/data.bin"
	httpGetStatusCode(t, urlPath, 404)

	// Invalid package version
	urlPath = "/files/foo/no-number/213151e5833fecb107899dfd0c8baca0fb671d4017fbd9361c8007b7b93681a6/data.bin"
	httpGetStatusCode(t, urlPath, 400)

	// Wrong hash
	urlPath = "/files/foo/1/666151e5833fecb107899dfd0c8baca0fb671d4017fbd9361c8007b7b9368666/data.bin"
	httpGetStatusCode(t, urlPath, 404)
}

func TestServerGzipFileHandler(t *testing.T) {
	// Prepare Cleanup
	defer os.RemoveAll(storeFolder)
	defer os.RemoveAll(packageFolderBig)

	// Creation and cleanup of server store
	server, stopped := startTestingServer(t)
	defer stopTestingServer(server, stopped)

	// Publish test package
	publishBigTestPackage(t)

	// Request file under test
	urlPath := "/files/bar/1/e22e4bb46ad3e963fe059dcd969c036bd556a020d1de2d8cbd393a19ee74eb8c/testfile.dat"
	body, headers, err := httpGet(urlPath)
	if err != nil {
		t.Fatal(err)
	}

	// File is big enough to trigger gzip compression
	if headers["Content-Encoding"][0] != "gzip" {
		t.Fatal(headers)
	}

	// File contains a number of sequential files and should compress well
	if len(headers) >= 1024 {
		t.Fatal()
	}

	// Decompress gzip encoded data
	buf := bytes.NewBuffer(body)
	reader, err := gzip.NewReader(buf)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	decompressed, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}

	// Check decompressed data for original length
	if len(decompressed) != 1024 {
		t.Fatal(len(decompressed))
	}
}

func TestServerZipHandler(t *testing.T) {
	// Prepare Cleanup
	defer os.RemoveAll(storeFolder)
	defer os.RemoveAll(unzipFolder)

	// Creation and cleanup of server store
	server, stopped := startTestingServer(t)
	defer stopTestingServer(server, stopped)

	// Publish test package
	publishSmallTestPackage(t)

	// Request ZIP of package
	urlPath := "/zip/foo/1"
	body, headers, err := httpGet(urlPath)
	if err != nil {
		t.Fatal(err)
	}

	// Content type must be ZIP
	if headers["Content-Type"][0] != "application/zip" {
		t.Fatal(headers)
	}

	// Download name must be <package>.<version>.zip
	if headers["Content-Disposition"][0] != "attachment; filename=\"foo.v1.zip\"" {
		t.Fatal(headers)
	}

	// Open returned body as ZIP archive
	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}

	// Read all files from ZIP
	for _, zipFile := range zipReader.File {
		file, err := zipFile.Open()
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		unzippedFileData, err := ioutil.ReadAll(file)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(unzipFolder, zipFile.Name)
		folder := filepath.Dir(path)
		err = os.MkdirAll(folder, os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}
		err = ioutil.WriteFile(path, unzippedFileData, os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Generate manifests for original and unzipped folder and compare
	manifestOrg, err := bdm.GenerateManifest(packageNameSmall, packageFolderSmall)
	if err != nil {
		t.Fatal(err)
	}
	manifestZipped, err := bdm.GenerateManifest(packageNameSmall, unzipFolder)
	if err != nil {
		t.Fatal(err)
	}
	if manifestOrg.Hash != manifestZipped.Hash {
		t.Fatal(manifestOrg, manifestZipped)
	}
}

func TestServerStaticHandler(t *testing.T) {
	// Creation and cleanup of server store
	server, stopped := startTestingServer(t)
	defer stopTestingServer(server, stopped)

	// Request UI
	urlPath := "/"
	body, headers, err := httpGet(urlPath)
	if err != nil {
		t.Fatal(err)
	}

	// Content type must be HTML
	if !strings.Contains(headers["Content-Type"][0], "text/html") {
		t.Fatal(headers)
	}

	// Check for HTML content
	if len(body) <= 0 || !strings.Contains(string(body), "<html") {
		t.Fatal(body)
	}

	// Request favicon
	urlPath = "/favicon.ico"
	body, headers, err = httpGet(urlPath)
	if err != nil {
		t.Fatal(err)
	}

	// Check content type
	if headers["Content-Type"][0] != "image/x-icon" {
		t.Fatal(headers)
	}

	if len(body) <= 0 {
		t.Fatal(body)
	}
}

func startTestingServer(t *testing.T) (*http.Server, chan bool) {
	packageStore, err := store.New(storeFolder)
	if err != nil {
		t.Fatal(err)
	}

	limits := bdm.ManifestLimits{}
	permissions := server.SimplePermissions("", token)
	handler, err := server.CreateRouter(packageStore, &limits, permissions)
	if err != nil {
		t.Fatal(err)
	}

	server := &http.Server{Addr: "127.0.0.1:2323", Handler: handler}
	stopped := make(chan bool)
	go func() {
		server.ListenAndServe()
		stopped <- true
	}()

	// Wait until server has started
	for {
		time.Sleep(time.Millisecond)
		resp, err := http.Get("http://127.0.0.1:2323/")
		if err == nil && resp.StatusCode == 200 {
			break
		}
	}

	return server, stopped
}

func stopTestingServer(server *http.Server, stopped chan bool) {
	server.Shutdown(context.Background())
	<-stopped // Wait for server to be stopped
}

func publishSmallTestPackage(t *testing.T) {
	manifest, err := client.UploadPackage(packageNameSmall, packageFolderSmall, serverURL, token)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.PackageVersion != 1 || manifest.Published == 0 || len(manifest.Files) <= 0 {
		t.Fatal(manifest)
	}
}

func generateTestFile(filePath string, size, seed int) error {
	handle, err := os.Create(filePath)
	if err != nil {
		return err
	}
	for i := 0; i < size; i++ {
		written, err := handle.Write([]byte{byte(seed)})
		seed = seed + 1
		if written != 1 {
			return fmt.Errorf("failed to generate test data")
		}
		if err != nil {
			return err
		}
	}
	return handle.Close()
}

func publishBigTestPackage(t *testing.T) {
	os.MkdirAll(packageFolderBig, os.ModePerm)
	err := generateTestFile(filepath.Join(packageFolderBig, "testfile.dat"), 1024, 666)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := client.UploadPackage(packageNameBig, packageFolderBig, serverURL, token)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.PackageVersion != 1 || manifest.Published == 0 || len(manifest.Files) <= 0 {
		t.Fatal(manifest)
	}
}

func httpGet(path string) ([]byte, http.Header, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://127.0.0.1:2323"+path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept-Encoding", "gzip")
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, nil, fmt.Errorf("returned status code %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	return body, resp.Header, nil
}

func httpGetStatusCode(t *testing.T, path string, statusCode int) {
	resp, err := http.Get("http://127.0.0.1:2323" + path)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != statusCode {
		t.Fatal(resp.StatusCode)
	}
}

func getAndCompareString(t *testing.T, path, expectedType, expectedStr string) {
	body, header, err := httpGet(path)
	if err != nil {
		t.Fatal(err)
	}
	if header["Content-Type"][0] != expectedType {
		t.Fatal(header)
	}
	str := string(body)
	if str != expectedStr {
		t.Fatal(str)
	}
}
