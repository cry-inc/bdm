package main

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
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

const storeFolder = "test/store"
const outputFolder = "test/output"
const cacheFolder = "test/cache"
const serverURL = "http://127.0.0.1:2323"
const packageNameSmall = "foo"
const packageFolderSmall = "test/example"
const packageNameBig = "bar"
const packageFolderBig = "test/big"
const unzipFolder = "test/unzipped"

// will be set later during test setup
var readToken string
var writeToken string

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
	err := client.DownloadPackage(outputFolder, serverURL, readToken, packageNameSmall, 1, false)
	util.AssertNoError(t, err)

	// Check package content
	err = client.CheckPackage(outputFolder, serverURL, readToken, packageNameSmall, 1, true)
	util.AssertNoError(t, err)

	// Create dirty file and download again with clean mode enabled
	const dirtyFile = outputFolder + "/dirty.dat"
	os.WriteFile(dirtyFile, []byte{0, 1, 2}, os.ModePerm)
	err = client.DownloadPackage(outputFolder, serverURL, readToken, packageNameSmall, 1, true)
	util.AssertNoError(t, err)
	util.Assert(t, !util.FileExists(dirtyFile))

	// Download again with caching enabled
	os.RemoveAll(outputFolder)
	err = client.DownloadCachedPackage(outputFolder, cacheFolder, serverURL, readToken, packageNameSmall, 1, false)
	util.AssertNoError(t, err)

	// Stop server
	stopTestingServer(server, stopped)

	// Try to restore package from cache only
	os.RemoveAll(outputFolder)
	err = client.DownloadCachedPackage(outputFolder, cacheFolder, serverURL, readToken, packageNameSmall, 1, true)
	util.AssertNoError(t, err)

	// Check using the cache only
	err = client.CheckCachedPackage(outputFolder, cacheFolder, serverURL, readToken, packageNameSmall, 1, true)
	util.AssertNoError(t, err)
}

func TestServerJsonHandlers(t *testing.T) {
	// Prepare Cleanup
	defer os.RemoveAll(storeFolder)

	// Creation and cleanup of server store
	server, stopped := startTestingServer(t)
	defer stopTestingServer(server, stopped)

	// Check empty manifest list on fresh server
	getAndCompareString(t, "/manifests", readToken, "application/json", "[]")

	// Publish a small test package
	publishSmallTestPackage(t)

	// Look for fresh test package in the list
	getAndCompareString(t, "/manifests", readToken, "application/json", "[{\"Name\":\"foo\"}]")

	// List versions of the fresh test package
	getAndCompareString(t, "/manifests/foo", readToken, "application/json", "[{\"Version\":1}]")
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
	getAndCompareString(t, urlPath, readToken, "application/octet-stream", expectedData)

	// Empty file
	urlPath = "/files/foo/1/af1349b9f5f9a1a6a0404dea36dcc9499bcb25c9adc112b7cc9a93cae41f3262/empty.txt"
	getAndCompareString(t, urlPath, readToken, "application/octet-stream", "")

	// Wrong file name
	urlPath = "/files/foo/1/213151e5833fecb107899dfd0c8baca0fb671d4017fbd9361c8007b7b93681a6/wrong.name"
	httpGetStatusCode(t, urlPath, readToken, 404)

	// Wrong package name
	urlPath = "/files/blaaaaa/1/213151e5833fecb107899dfd0c8baca0fb671d4017fbd9361c8007b7b93681a6/data.bin"
	httpGetStatusCode(t, urlPath, readToken, 404)

	// Invalid package name
	urlPath = "/files/bla+()aaaa/1/213151e5833fecb107899dfd0c8baca0fb671d4017fbd9361c8007b7b93681a6/data.bin"
	httpGetStatusCode(t, urlPath, readToken, 400)

	// Wrong package version
	urlPath = "/files/foo/666/213151e5833fecb107899dfd0c8baca0fb671d4017fbd9361c8007b7b93681a6/data.bin"
	httpGetStatusCode(t, urlPath, readToken, 404)

	// Invalid package version
	urlPath = "/files/foo/no-number/213151e5833fecb107899dfd0c8baca0fb671d4017fbd9361c8007b7b93681a6/data.bin"
	httpGetStatusCode(t, urlPath, readToken, 400)

	// Wrong hash
	urlPath = "/files/foo/1/666151e5833fecb107899dfd0c8baca0fb671d4017fbd9361c8007b7b9368666/data.bin"
	httpGetStatusCode(t, urlPath, readToken, 404)
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
	body, headers, err := httpGet(urlPath, readToken)
	util.AssertNoError(t, err)

	// File is big enough to trigger gzip compression
	util.AssertEqualString(t, "gzip", headers["Content-Encoding"][0])

	// File contains a number of sequential files and should compress well
	util.Assert(t, len(body) < 1024)

	// Decompress gzip encoded data
	buf := bytes.NewBuffer(body)
	reader, err := gzip.NewReader(buf)
	util.AssertNoError(t, err)
	defer reader.Close()
	decompressed, err := io.ReadAll(reader)
	util.AssertNoError(t, err)

	// Check decompressed data for original length
	util.Assert(t, len(decompressed) == 1024)
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
	body, headers, err := httpGet(urlPath, readToken)
	util.AssertNoError(t, err)

	// Content type must be ZIP
	util.AssertEqualString(t, "application/zip", headers["Content-Type"][0])

	// Download name must be <package>.<version>.zip
	util.AssertEqualString(t, "attachment; filename=\"foo.v1.zip\"", headers["Content-Disposition"][0])

	// Open returned body as ZIP archive
	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	util.AssertNoError(t, err)

	// Read all files from ZIP
	for _, zipFile := range zipReader.File {
		file, err := zipFile.Open()
		util.AssertNoError(t, err)
		defer file.Close()
		unzippedFileData, err := io.ReadAll(file)
		util.AssertNoError(t, err)
		path := filepath.Join(unzipFolder, zipFile.Name)
		folder := filepath.Dir(path)
		err = os.MkdirAll(folder, os.ModePerm)
		util.AssertNoError(t, err)
		err = os.WriteFile(path, unzippedFileData, os.ModePerm)
		util.AssertNoError(t, err)
	}

	// Generate manifests for original and unzipped folder and compare
	manifestOrg, err := bdm.GenerateManifest(packageNameSmall, packageFolderSmall)
	util.AssertNoError(t, err)
	manifestZipped, err := bdm.GenerateManifest(packageNameSmall, unzipFolder)
	util.AssertNoError(t, err)
	util.AssertEqualString(t, manifestOrg.Hash, manifestZipped.Hash)
}

func TestServerStaticHandler(t *testing.T) {
	// Creation and cleanup of server store
	server, stopped := startTestingServer(t)
	defer stopTestingServer(server, stopped)

	// Request UI
	urlPath := "/"
	body, headers, err := httpGet(urlPath, readToken)
	util.AssertNoError(t, err)

	// Content type must be HTML
	util.Assert(t, strings.Contains(string(headers["Content-Type"][0]), "text/html"))

	// Check for HTML content
	util.Assert(t, len(body) > 0)
	util.Assert(t, strings.Contains(string(body), "</html>"))

	// Request favicon
	urlPath = "/favicon.ico"
	body, headers, err = httpGet(urlPath, readToken)
	util.AssertNoError(t, err)

	// Check content type
	mimeType := headers["Content-Type"][0]
	// Since Go also checks the OS type lists we might get different results on different systems
	util.Assert(t, mimeType == "image/x-icon" || mimeType == "image/vnd.microsoft.icon")
	util.Assert(t, len(body) > 0)
}

func startTestingServer(t *testing.T) (*http.Server, chan bool) {
	packageStore, err := store.New(storeFolder)
	util.AssertNoError(t, err)

	limits := bdm.ManifestLimits{}
	users, err := server.CreateJsonUsers("./users.json")
	util.AssertNoError(t, err)
	defer os.Remove("./users.json")
	tokens, err := server.CreateJsonTokens("./tokens.json", users, false, false)
	util.AssertNoError(t, err)
	defer os.Remove("./tokens.json")
	handler := server.CreateRouter(packageStore, &limits, users, tokens)

	err = users.CreateUser(server.User{
		Id: "admin",
		Roles: server.Roles{
			Admin:  true,
			Writer: true,
			Reader: true,
		},
	}, "mypassword")
	util.AssertNoError(t, err)

	rt, err := tokens.CreateToken("admin", &server.Roles{Reader: true})
	util.AssertNoError(t, err)
	readToken = rt.Id

	wt, err := tokens.CreateToken("admin", &server.Roles{Writer: true, Reader: true})
	util.AssertNoError(t, err)
	writeToken = wt.Id

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
	manifest, err := client.UploadPackage(packageNameSmall, packageFolderSmall, serverURL, writeToken)
	util.AssertNoError(t, err)
	util.Assert(t, manifest.PackageVersion == 1)
	util.Assert(t, manifest.Published != 0)
	util.Assert(t, len(manifest.Files) > 0)
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
	util.AssertNoError(t, err)
	manifest, err := client.UploadPackage(packageNameBig, packageFolderBig, serverURL, writeToken)
	util.AssertNoError(t, err)
	util.Assert(t, manifest.PackageVersion == 1)
	util.Assert(t, manifest.Published != 0)
	util.Assert(t, len(manifest.Files) > 0)
}

func httpGet(path, token string) ([]byte, http.Header, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://127.0.0.1:2323"+path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set(bdm.ApiTokenHeader, token)
	req.Header.Set("Accept-Encoding", "gzip")
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, nil, fmt.Errorf("returned status code %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	return body, resp.Header, nil
}

func httpGetStatusCode(t *testing.T, path, token string, statusCode int) {
	t.Helper()
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://127.0.0.1:2323"+path, nil)
	util.AssertNoError(t, err)
	req.Header.Set(bdm.ApiTokenHeader, token)
	resp, err := client.Do(req)
	util.AssertNoError(t, err)
	defer resp.Body.Close()
	util.Assert(t, resp.StatusCode == statusCode)
}

func getAndCompareString(t *testing.T, path, token, expectedType, expectedStr string) {
	t.Helper()
	body, header, err := httpGet(path, token)
	util.AssertNoError(t, err)
	util.AssertEqualString(t, expectedType, header["Content-Type"][0])
	util.AssertEqualString(t, expectedStr, string(body))
}
