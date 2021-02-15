package util

import (
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
)

// CreateCompressingWriter returns a compressing writer
func CreateCompressingWriter(writer io.Writer) (io.WriteCloser, error) {
	options := zstd.WithEncoderLevel(zstd.SpeedDefault)
	return zstd.NewWriter(writer, options)
}

// CreateDecompressingReader returns a decompressing reader
func CreateDecompressingReader(reader io.Reader) (io.ReadCloser, error) {
	decoder, err := zstd.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("error creating zstd reader: %w", err)
	}

	// We cannot directly return the decoder as ReadCloser since it has
	// the wrong Close() signature. This can be fixed by wrapping it.
	var zrc zstdReadCloser
	zrc.reader = decoder

	return zrc, nil
}

type zstdReadCloser struct {
	reader *zstd.Decoder
}

func (zrc zstdReadCloser) Close() error {
	zrc.reader.Close() // does not return an error :(
	return nil
}

func (zrc zstdReadCloser) Read(p []byte) (int, error) {
	return zrc.reader.Read(p)
}
