package util

import (
	"io"
	"io/ioutil"
	"testing"
)

func TestCompression(t *testing.T) {
	testData := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 10, 11, 12, 13, 14, 15, 16}

	r, w := io.Pipe()
	cr, err := CreateDecompressingReader(r)
	if err != nil {
		t.Fatal(err)
	}
	defer cr.Close()

	cw, err := CreateCompressingWriter(w)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		defer w.Close()
		defer cw.Close()

		n, err := cw.Write(testData)
		if err != nil {
			panic(err)
		}
		if n != len(testData) {
			panic("Failed to write data")
		}
	}()

	readData, err := ioutil.ReadAll(cr)
	if err != nil {
		t.Fatal(err)
	}

	if len(readData) != len(testData) {
		t.Fatal()
	}

	for i := range readData {
		if readData[i] != testData[i] {
			t.Fatal(i)
		}
	}
}
