package bdm

import (
	"io"
	"testing"

	"github.com/cry-inc/bdm/pkg/bdm/util"
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
	util.AssertNoError(t, err)
	util.Assert(t, len(read) == len(objects))
	util.Assert(t, read[0] == objects[0])
	util.Assert(t, read[1] == objects[1])
}
