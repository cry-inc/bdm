package bdm

import (
	"fmt"
	"io"
	"reflect"
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
			panic(fmt.Errorf("failed to write objects to stream: %w", err))
		}
	}()

	read, err := ReadObjectsFromStream(reader)
	util.AssertNoError(t, err)
	util.Assert(t, reflect.DeepEqual(read, objects))
}
