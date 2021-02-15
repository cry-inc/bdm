package main

import "C"

import (
	"git.caputo.de/macaputo/bdm/pkg/bdm/client"
)

func main() {}

// Downloads a remote package to the specified output folder.
// Will remove all non-package files from the output folder of clean is set to a non-zero value.
// Return value will be zero when successful.
//export bdmDownloadPackage
func bdmDownloadPackage(packageName *C.char, packageVersion C.int, outputFolder, serverURL *C.char, clean C.int) C.int {
	goPackageName := C.GoString(packageName)
	goOutputFolder := C.GoString(outputFolder)
	goServerURL := C.GoString(serverURL)

	err := client.DownloadPackage(goOutputFolder, goServerURL, goPackageName, uint(packageVersion), clean != 0)
	if err != nil {
		return 1
	}

	return 0
}

// Like bdmDownloadPackage with an additional local persistent cache in the specified cache folder.
//export bdmDownloadCachedPackage
func bdmDownloadCachedPackage(packageName *C.char, cacheFolder *C.char, packageVersion C.int, outputFolder, serverURL *C.char, clean C.int) C.int {
	goPackageName := C.GoString(packageName)
	goOutputFolder := C.GoString(outputFolder)
	goCacheFolder := C.GoString(cacheFolder)
	goServerURL := C.GoString(serverURL)

	err := client.DownloadCachedPackage(goOutputFolder, goCacheFolder, goServerURL, goPackageName, uint(packageVersion), clean != 0)
	if err != nil {
		return 1
	}

	return 0
}

// Checks if the content of a local folder matches the specified package from a server.
// Will report errors for any non-package files in the folder in case clean is non-zero.
// Return value will be zero when successful.
//export bdmCheckPackage
func bdmCheckPackage(packageName *C.char, packageVersion C.int, packageFolder, serverURL *C.char, clean C.int) C.int {
	goPackageName := C.GoString(packageName)
	goPackageFolder := C.GoString(packageFolder)
	goServerURL := C.GoString(serverURL)

	err := client.CheckPackage(goPackageFolder, goServerURL, goPackageName, uint(packageVersion), clean != 0)
	if err != nil {
		return 1
	}

	return 0
}

// Like bdmCheckPackage with an additional local persistent cache in the specified cache folder.
//export bdmCheckCachedPackage
func bdmCheckCachedPackage(packageName *C.char, cacheFolder *C.char, packageVersion C.int, packageFolder, serverURL *C.char, clean C.int) C.int {
	goPackageName := C.GoString(packageName)
	goPackageFolder := C.GoString(packageFolder)
	goCacheFolder := C.GoString(cacheFolder)
	goServerURL := C.GoString(serverURL)

	err := client.CheckCachedPackage(goPackageFolder, goCacheFolder, goServerURL, goPackageName, uint(packageVersion), clean != 0)
	if err != nil {
		return 1
	}

	return 0
}

// Uploads and publishes the files from the specified local folder as package on the remote server.
// After successful publishing the output argument version will contain the new version number assigned by the server.
// Return value will be zero when successful.
//export bdmUploadPackage
func bdmUploadPackage(packageName *C.char, packageFolder *C.char, serverURL, apiKey *C.char, version *C.int) C.int {
	goPackageName := C.GoString(packageName)
	goPackageFolder := C.GoString(packageFolder)
	goServerURL := C.GoString(serverURL)
	goAPIKey := C.GoString(apiKey)

	manifest, err := client.UploadPackage(goPackageName, goPackageFolder, goServerURL, goAPIKey)
	if err != nil {
		return 1
	}

	*version = C.int(manifest.PackageVersion)

	return 0
}
