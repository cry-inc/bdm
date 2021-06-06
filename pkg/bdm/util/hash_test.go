package util

import (
	"bytes"
	"os"
	"testing"
)

func TestCreateHasher(t *testing.T) {
	hasher := CreateHasher()
	sum := hasher.Sum(nil)
	Assert(t, len(sum) > 0)
}

func TestHashStream(t *testing.T) {
	buf := bytes.NewBuffer([]byte{1, 2, 3, 4})
	hash, err := HashStream(buf)
	AssertNoError(t, err)
	AssertEqualString(t, "63781d171425a36312fa058d8712d5d05135a991ec20351ce9d65cdb19a05432", hash)
}

func TestHashFile(t *testing.T) {
	hash, err := HashFile("filedoesnotexist")
	AssertError(t, err)
	AssertEqualString(t, "", hash)

	testFile := "test.dat"
	testData := []byte{0, 1, 2, 3, 4, 5, 6}
	err = os.WriteFile(testFile, testData, os.ModePerm)
	AssertNoError(t, err)
	hash, err = HashFile(testFile)
	AssertNoError(t, err)
	AssertEqualString(t, "3f8770f387faad08faa9d8414e9f449ac68e6ff0417f673f602a646a891419fe", hash)
	err = os.Remove(testFile)
	AssertNoError(t, err)
}
