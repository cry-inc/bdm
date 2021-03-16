package bdm

import "fmt"

// ManifestLimits represents constraint for packages.
// The dfault value zero means unlimited.
type ManifestLimits struct {
	MaxFileSize    int64
	MaxPackageSize int64
	MaxFilesCount  int
	MaxPathLength  int
}

// CheckManifestLimits can check if a manifest is within the given package limmits.
// It will return nil if the manifest is within the limits, otherwise an error.
func CheckManifestLimits(manifest *Manifest, limits *ManifestLimits) error {
	if limits.MaxFilesCount > 0 && len(manifest.Files) > limits.MaxFilesCount {
		return fmt.Errorf("number of files is %d and exceeds the limit of %d",
			len(manifest.Files), limits.MaxFilesCount)
	}

	var overallSize int64 = 0
	for _, file := range manifest.Files {
		overallSize += file.Object.Size
		if limits.MaxPathLength > 0 && len(file.Path) > limits.MaxPathLength {
			return fmt.Errorf("path length of %d exceeds the limit of %d",
				len(file.Path), limits.MaxPathLength)
		}
		if limits.MaxFileSize > 0 && file.Object.Size > limits.MaxFileSize {
			return fmt.Errorf("file size of %d exceeds the limit of %d",
				file.Object.Size, limits.MaxFileSize)
		}
	}

	if limits.MaxPackageSize > 0 && overallSize > limits.MaxPackageSize {
		return fmt.Errorf("package size of %d exceeds the limit of %d",
			overallSize, limits.MaxPackageSize)
	}

	return nil
}
