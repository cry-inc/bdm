package bdm

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func checkName(t *testing.T, name string, valid bool) {
	if ValidatePackageName(name) != valid {
		t.Fatal(name)
	}
}

func TestValidatePackageName(t *testing.T) {
	checkName(t, "abc123", true)
	checkName(t, "abc-123_def", true)

	checkName(t, "ABC123", false)
	checkName(t, "abc-123_def.a", false)
	checkName(t, "äöüß", false)
}

func generateUnpublishedManifest() Manifest {
	manifest := Manifest{
		ManifestVersion: 1,
		PackageName:     "s123",
		Files: []File{
			{
				Path: "folder/scooby.doo",
				Object: Object{
					Size: 123,
					Hash: "abc0123456789def",
				},
			},
		},
	}
	manifest.Hash = HashManifest(&manifest)
	return manifest
}

func checkUnpublishedManifest(t *testing.T, manifest *Manifest, valid bool) {
	err := ValidateUnpublishedManifest(manifest)
	if valid && err != nil || !valid && err == nil {
		t.Fatal(manifest)
	}
}

func TestValidateUnpublishedManifest(t *testing.T) {
	manifest := generateUnpublishedManifest()
	checkUnpublishedManifest(t, &manifest, true)

	manifest.PackageVersion = 1
	checkUnpublishedManifest(t, &manifest, false)

	manifest = generateUnpublishedManifest()
	manifest.PackageName = "ABC"
	checkUnpublishedManifest(t, &manifest, false)

	manifest = generateUnpublishedManifest()
	manifest.Files = make([]File, 0)
	checkUnpublishedManifest(t, &manifest, false)

	manifest = generateUnpublishedManifest()
	manifest.Files[0].Path = ""
	checkUnpublishedManifest(t, &manifest, false)

	manifest = generateUnpublishedManifest()
	manifest.Files[0].Object.Hash = "ztgf/)"
	checkUnpublishedManifest(t, &manifest, false)

	manifest = generateUnpublishedManifest()
	manifest.Files[0].Path = "../" + manifest.Files[0].Path
	checkUnpublishedManifest(t, &manifest, false)

	manifest = generateUnpublishedManifest()
	manifest.Files[0].Object.Size = -1
	checkUnpublishedManifest(t, &manifest, false)

	manifest = generateUnpublishedManifest()
	duplicate := File{
		Path: "FOLDER/Scooby.DOO",
		Object: Object{
			Size: 456,
			Hash: "dcef01234",
		},
	}
	manifest.Files = append(manifest.Files, duplicate)
	checkUnpublishedManifest(t, &manifest, false)
}

func checkPublishedManifest(t *testing.T, manifest *Manifest, valid bool) {
	err := ValidatePublishedManifest(manifest)
	if valid && err != nil || !valid && err == nil {
		t.Fatal(manifest)
	}
}

func TestValidatePublishedManifest(t *testing.T) {
	manifest := generateUnpublishedManifest()
	checkPublishedManifest(t, &manifest, false)

	manifest = generateUnpublishedManifest()
	manifest.PackageVersion = 1
	manifest.Published = time.Now().Unix()
	manifest.Hash = HashManifest(&manifest)
	checkPublishedManifest(t, &manifest, true)
}

func TestGenerateManifest(t *testing.T) {
	testFolder := "testPackage"
	err := os.MkdirAll(testFolder+"/dir/subdir", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	err = ioutil.WriteFile(testFolder+"/dir/subdir/my file äöü 人物.txt", []byte{1, 2, 3}, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	manifest, err := GenerateManifest("foo", testFolder)
	if err != nil || manifest == nil {
		t.Fatal()
	}

	if manifest.PackageName != "foo" || manifest.PackageVersion != 0 || len(manifest.Files) != 1 {
		t.Fatal()
	}

	if manifest.Files[0].Path != "dir/subdir/my file äöü 人物.txt" ||
		manifest.Files[0].Object.Size != 3 ||
		manifest.Files[0].Object.Hash != "b177ec1bf26dfb3b7010d473e6d44713b29b765b99c6e60ecbfae742de496543" {
		t.Fatal(manifest.Files[0])
	}

	err = ValidateUnpublishedManifest(manifest)
	if err != nil {
		t.Fatal(err)
	}

	err = os.RemoveAll(testFolder)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHashManifest(t *testing.T) {
	emptyManifest := Manifest{}
	hash := HashManifest(&emptyManifest)
	if hash != "381507c20d3226db750821ad83686480a8ea69f56784598587161be18995170f" {
		t.Fatal(hash)
	}

	unpublishedManifest := generateUnpublishedManifest()
	hash = HashManifest(&unpublishedManifest)
	if hash != "b17a4f899214aac5fcb4924e3e9fe005062baeaf78ea781a1ac39aab76ea07c6" {
		t.Fatal(hash)
	}

	publishedManifest := unpublishedManifest
	publishedManifest.PackageVersion = 1
	publishedManifest.Published = 123456
	hash = HashManifest(&publishedManifest)
	if hash != "db910b1dba2bf0dc19247346622c3f9f14c8719eda01ea41b71cfaf13626dce2" {
		t.Fatal(hash)
	}
}
