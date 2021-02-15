package util

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestCreateHasher(t *testing.T) {
	hasher := CreateHasher()
	sum := hasher.Sum(nil)
	if len(sum) <= 0 {
		t.Fatal()
	}
}

func TestHashStream(t *testing.T) {
	buf := bytes.NewBuffer([]byte{1, 2, 3, 4})
	hash, err := HashStream(buf)
	if err != nil {
		t.Fatal(err)
	}
	if hash != "63781d171425a36312fa058d8712d5d05135a991ec20351ce9d65cdb19a05432" {
		t.Fatal(hash)
	}
}

func TestHashFile(t *testing.T) {
	hash, err := HashFile("filedoesnotexist")
	if err == nil || hash != "" {
		t.Fatal()
	}

	testFile := "test.dat"
	testData := []byte{0, 1, 2, 3, 4, 5, 6}
	err = ioutil.WriteFile(testFile, testData, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	hash, err = HashFile(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if hash != "3f8770f387faad08faa9d8414e9f449ac68e6ff0417f673f602a646a891419fe" {
		t.Fatal(hash)
	}
	err = os.Remove(testFile)
	if err != nil {
		t.Fatal(err)
	}
}
