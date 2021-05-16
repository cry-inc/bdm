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
	AssertNoError(t, err)
	defer cr.Close()

	cw, err := CreateCompressingWriter(w)
	AssertNoError(t, err)

	go func() {
		defer w.Close()
		defer cw.Close()

		n, err := cw.Write(testData)
		if err != nil {
			panic(err)
		}
		if n != len(testData) {
			panic("failed to write data")
		}
	}()

	readData, err := ioutil.ReadAll(cr)
	AssertNoError(t, err)
	Assert(t, len(readData) == len(testData))

	for i := range readData {
		Assert(t, readData[i] == testData[i])
	}
}
