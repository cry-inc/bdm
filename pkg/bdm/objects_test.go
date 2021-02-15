package bdm

import (
	"io"
	"testing"
)

func TestObjects(t *testing.T) {
	objects := []Object{
		{
			Hash: "abc123",
			Size: 123,
		},
		{
			Hash: "def456",
			Size: 456,
		},
	}

	reader, writer := io.Pipe()
	go func() {
		defer writer.Close()
		err := WriteObjectsToStream(objects, writer)
		if err != nil {
			panic(err)
		}
	}()

	read, err := ReadObjectsFromStream(reader)
	if err != nil {
		t.Fatal(err)
	}

	if len(read) != len(objects) || read[0] != objects[0] || read[1] != objects[1] {
		t.Fatal(read)
	}
}
