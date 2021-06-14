package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"runtime"

	"github.com/cry-inc/bdm/pkg/bdm"
	"github.com/cry-inc/bdm/pkg/bdm/client"
	"github.com/cry-inc/bdm/pkg/bdm/server"
	"github.com/cry-inc/bdm/pkg/bdm/store"
	"github.com/cry-inc/bdm/pkg/bdm/util"
)

// BDM version (optionally injected during build)
var bdmVersion = "n/a"

// git commit hash of build (optionally injected during build)
var commitHash = "n/a"

// Date string of this build (optionally injected during build)
var buildDate = "n/a"

func main() {
	// Application Modes
	uploadMode := flag.Bool("upload", false, "Enables upload mode to publish new packages.")
	downloadMode := flag.Bool("download", false, "Enables download mode to get remote packages.")
	serverMode := flag.Bool("server", false, "Enables server mode to run a package repository server.")
	checkMode := flag.Bool("check", false, "Enables check mode to compare local folder against an existing package.")
	aboutMode := flag.Bool("about", false, "Show application version and build information.")
	validateMode := flag.Bool("validate", false, "Validates a package store to make sure all contained data is valid.")

	// Application Arguments
	port := flag.Uint("port", 2323, "Port for HTTP server of the package repository in server mode.")
	token := flag.String("token", "", "API token used for authorization in client mode.")
	httpsCert := flag.String("httpscert", "", "If supplied together with httpskey this will enable HTTPS.")
	httpsKey := flag.String("httpskey", "", "If supplied together with httpscert this will enable HTTPS.")
	letsEncryptDomain := flag.String("letsencrypt", "", "Domain name to enable HTTPS with automatic LE certificates. Will also start an HTTP server on port 80 that needs to be reachable from the internet.")
	certCacheFolder := flag.String("certcache", "./certs", "Cache folder for LE certificates.")
	storeFolder := flag.String("store", "./store", "Specifies location of the servers package repository on disk.")
	guestReading := flag.Bool("guestreading", false, "Use this flag to allow everyone without an account to browse and download packages.")
	guestWriting := flag.Bool("guestwriting", false, "Use this flag to allow everyone without an account to upload new packages. Not recommended!")
	usersFile := flag.String("usersfile", "./users.json", "Specifies location of the servers JSON user database.")
	tokensFile := flag.String("tokensfile", "./tokens.json", "Specifies location of the servers JSON tokens database.")
	defaultUser := flag.String("defaultuser", "admin", "Specifies the name of the first user that will be automatically generated.")
	packageVersion := flag.Uint("version", 0, "Package version to download or check.")
	packageName := flag.String("package", "", "Specifies name of the package to be uploaded, downloaded or checked.")
	inputFolder := flag.String("input", "", "Input path to folder that contains the package data to be published or checked.")
	outputFolder := flag.String("output", "", "Output path to folder that receives the downloaded package data.")
	remoteServer := flag.String("remote", "", "Remote package server URL for downloading packages.")
	cacheFolder := flag.String("cache", "", "Local cache folder to avoid re-downloading packages from a remote server.")
	clean := flag.Bool("clean", false, "Deletes all non-package files in the output folder in download mode and ensures that there are no non-package files in check mode.")
	maxPathLength := flag.Int("maxpath", 0, "Maximum length of paths inside packages. Default is 0, which means unlimited.")
	maxFileCount := flag.Int("maxfiles", 0, "Maximum bumber of files per package. Default is 0, which means unlimited.")
	maxPackageSize := flag.Int64("maxsize", 0, "Maximum package size (sum of file sizes) in bytes. Default is 0, which means unlimited.")
	maxFileSize := flag.Int64("maxfilesize", 0, "Maximum file size inside packages in bytes. Default is 0, which means unlimited.")

	flag.Parse()

	limits := bdm.ManifestLimits{
		MaxFileSize:    *maxFileSize,
		MaxPackageSize: *maxPackageSize,
		MaxFilesCount:  *maxFileCount,
		MaxPathLength:  *maxPathLength,
	}

	if *serverMode {
		startServer(*port, &limits, *storeFolder, *usersFile, *defaultUser, *tokensFile, *guestReading, *guestWriting, *httpsCert, *httpsKey, *letsEncryptDomain, *certCacheFolder)
	} else if *validateMode {
		validateStore(*storeFolder)
	} else if *uploadMode {
		uploadPackage(*packageName, *inputFolder, *remoteServer, *token)
	} else if *downloadMode {
		downloadPackage(*packageName, *packageVersion, *outputFolder, *remoteServer, *token, *cacheFolder, *clean)
	} else if *checkMode {
		checkPackage(*packageName, *packageVersion, *inputFolder, *cacheFolder, *remoteServer, *token, *clean)
	} else if *aboutMode {
		showAbout()
	} else {
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func showAbout() {
	fmt.Printf("Build Info:\n")
	fmt.Printf("  Version: %s\n", bdmVersion)
	fmt.Printf("  Commit:  %s\n", commitHash)
	fmt.Printf("  Date:    %s\n", buildDate)
	fmt.Printf("  Go:      %s\n", runtime.Version())
	fmt.Printf("  OS:      %s\n", runtime.GOOS)
	fmt.Printf("  Arch:    %s\n", runtime.GOARCH)
}

func startServer(port uint, limits *bdm.ManifestLimits, storePath, usersFile, defaultUser, tokensFile string, guestReading, guestWriting bool, certPath, keyPath, letsEncryptDomain, certCacheFolder string) {
	if port == 0 || float64(port) >= math.Pow(2, 16) {
		log.Fatal("Invalid port number")
	}

	packageStore, err := store.New(storePath)
	if err != nil {
		log.Fatalf("Failed to open or create package store: %v", err)
	}

	users, err := server.CreateJsonUsers(usersFile)
	if err != nil {
		log.Fatalf("Failed to open or create user database: %v", err)
	}

	userList, err := users.GetUsers()
	if err != nil {
		log.Fatalf("Failed to get list of existing users: %v", err)
	}

	// Create default user if there are no users
	if len(userList) == 0 {
		password := util.GenerateRandomHexString(8)
		err = users.CreateUser(server.User{
			Id: defaultUser,
			Roles: server.Roles{
				Admin:  true,
				Writer: true,
				Reader: true,
			},
		}, password)
		if err != nil {
			log.Fatalf("Failed to create default user: %v", err)
		}
		fmt.Printf("Created default user '%s' with password '%s'\n", defaultUser, password)
	}

	tokens, err := server.CreateJsonTokens(tokensFile, users, guestReading, guestWriting)
	if err != nil {
		log.Fatalf("Failed to open or create token database: %v", err)
	}

	if guestWriting {
		fmt.Println("WARNING: Guest upload of new packages is enabled. This is not recommended!")
	}

	router := server.CreateRouter(packageStore, limits, users, tokens)

	p := uint16(port)
	if len(letsEncryptDomain) > 0 {
		fmt.Printf("Starting Let's Encrypt HTTPS server for domain %s on port %d and 80, cert cache folder '%s' and store folder '%s'\n",
			letsEncryptDomain, port, certCacheFolder, storePath)
		server.StartServerLetsEncrypt(p, letsEncryptDomain, certCacheFolder, router)
	} else if len(certPath) > 0 && len(keyPath) > 0 {
		fmt.Printf("Starting HTTPS server on port %d, with cert file '%s', key file '%s' and store folder '%s'\n",
			port, certPath, keyPath, storePath)
		server.StartServerTLS(p, certPath, keyPath, router)
	} else {
		fmt.Printf("Starting HTTP server on port %d and store folder '%s'\n", port, storePath)
		server.StartServer(p, router)
	}
}

func uploadPackage(packageName, inputFolder, serverURL, apiToken string) {
	validName := bdm.ValidatePackageName(packageName)
	if !validName {
		log.Fatal("Invalid package name. Only lower case a-z, 0-9 and the characters - _ are allowed")
	}

	if len(inputFolder) == 0 {
		log.Fatal("Missing input folder!")
	}

	if !util.FolderExists(inputFolder) {
		log.Fatal("Input folder does not exist")
	}

	err := validateServerURL(serverURL)
	if err != nil {
		log.Fatal(err)
	}

	manifest, err := client.UploadPackage(packageName, inputFolder, serverURL, apiToken)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Package %s was successfully published in version %d", manifest.PackageName, manifest.PackageVersion)
}

func downloadPackage(packageName string, packageVersion uint, outputFolder, serverURL, apiToken, cacheFolder string, clean bool) {
	if len(packageName) == 0 {
		log.Fatal("Missing package name")
	}

	if len(outputFolder) == 0 {
		log.Fatal("Missing output folder")
	}

	if packageVersion == 0 {
		log.Fatal("Missing or invalid package version")
	}

	err := validateServerURL(serverURL)
	if err != nil {
		log.Fatal(err)
	}

	if len(cacheFolder) > 0 {
		err = client.DownloadCachedPackage(outputFolder, cacheFolder, serverURL, apiToken, packageName, packageVersion, clean)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err = client.DownloadPackage(outputFolder, serverURL, apiToken, packageName, packageVersion, clean)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func checkPackage(packageName string, packageVersion uint, checkFolder, cacheFolder, serverURL, apiToken string, clean bool) {
	if len(packageName) == 0 {
		log.Fatal("Missing package name")
	}

	if packageVersion == 0 {
		log.Fatal("Missing or invalid package version")
	}

	err := validateServerURL(serverURL)
	if err != nil {
		log.Fatal(err)
	}

	if len(cacheFolder) > 0 {
		err = client.CheckCachedPackage(checkFolder, cacheFolder, serverURL, apiToken, packageName, packageVersion, clean)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err = client.CheckPackage(checkFolder, serverURL, apiToken, packageName, packageVersion, clean)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func validateStore(storeFolder string) {
	if !util.FolderExists(storeFolder) {
		log.Fatal("Missing store folder")
	}

	packageStore, err := store.New((storeFolder))
	if err != nil {
		log.Fatal(err)
	}

	stats, err := store.ValidateStore(packageStore)
	if err != nil {
		log.Fatal(err)
	}

	for name, value := range stats {
		fmt.Printf("%s: %d\n", name, value)
	}
}

func validateServerURL(serverURL string) error {
	if len(serverURL) == 0 {
		return fmt.Errorf("missing URL for remote server")
	}

	matched, _ := regexp.MatchString(`^http([s]?)://.+[^/]$`, serverURL)
	if !matched {
		return fmt.Errorf("server URL must be a valid HTTP or HTTPS URL without a trailing slash")
	}

	return nil
}
