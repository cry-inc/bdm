package server

import (
	"fmt"
	"io"

	"github.com/cry-inc/bdm/pkg/bdm"
	"github.com/cry-inc/bdm/pkg/bdm/store"
	"github.com/cry-inc/bdm/pkg/bdm/util"
)

func streamObjectsToStore(input io.Reader, store store.Store, maxObjectSize int64) ([]bdm.Object, error) {
	decompressedInput, err := util.CreateDecompressingReader(input)
	if err != nil {
		return nil, fmt.Errorf("error creating decompressing reader: %w", err)
	}
	defer decompressedInput.Close()

	objects, err := bdm.ReadObjectsFromStream(decompressedInput)
	if err != nil {
		return nil, fmt.Errorf("error reading objects from stream: %w", err)
	}

	addedObjects := make([]bdm.Object, 0)
	for _, object := range objects {
		if maxObjectSize > 0 && object.Size > maxObjectSize {
			return nil, fmt.Errorf("object size of %d exceeds limit of %d",
				object.Size, maxObjectSize)
		}

		reader, writer := io.Pipe()
		go func() {
			defer writer.Close()
			copied, err := io.CopyN(writer, decompressedInput, object.Size)
			if err != nil {
				panic(fmt.Errorf("failed to copy decompressed input data: %w", err))
			}
			if copied != object.Size {
				panic(fmt.Errorf("object size mismatch: expected %d but copied %d bytes", object.Size, copied))
			}
		}()

		addedObject, err := store.AddObject(reader)
		if err != nil {
			return nil, fmt.Errorf("error adding object %s to store: %w", object.Hash, err)
		}
		if addedObject.Hash != object.Hash {
			return nil, fmt.Errorf("object hash mismatch detected: expected %s but found %s",
				object.Hash, addedObject.Hash)
		}
		if addedObject.Size != object.Size {
			return nil, fmt.Errorf("object size mismatch detected: expected %d but found %d bytes",
				object.Size, addedObject.Size)
		}
		addedObjects = append(addedObjects, *addedObject)
	}

	return addedObjects, nil
}

func streamObjectsFromStore(input io.Reader, store store.Store, output io.Writer) error {
	decompressedInput, err := util.CreateDecompressingReader(input)
	if err != nil {
		return fmt.Errorf("error creating decompressing reader: %w", err)
	}
	defer decompressedInput.Close()

	objects, err := bdm.ReadObjectsFromStream(decompressedInput)
	if err != nil {
		return fmt.Errorf("error reading objects from stream: %w", err)
	}

	foundObjects := make([]bdm.Object, 0)
	for _, object := range objects {
		foundObject, err := store.GetObject(object.Hash)
		if err == nil &&
			foundObject.Size == object.Size &&
			foundObject.Hash == object.Hash {
			foundObjects = append(foundObjects, *foundObject)
		}
	}

	compressedOutput, err := util.CreateCompressingWriter(output)
	if err != nil {
		return fmt.Errorf("error creating compressing writer: %w", err)
	}
	defer compressedOutput.Close()

	err = bdm.WriteObjectsToStream(foundObjects, compressedOutput)
	if err != nil {
		return fmt.Errorf("error writing objects to stream: %w", err)
	}

	for _, object := range foundObjects {
		reader, err := store.ReadObject(object.Hash)
		if err != nil {
			return fmt.Errorf("error reading object %s from store: %w", object.Hash, err)
		}
		written, err := io.Copy(compressedOutput, reader)
		if err != nil {
			return fmt.Errorf("error copying object data: %w", err)
		}
		if written != object.Size {
			return fmt.Errorf("error copying object data: expected to write %d but wrote %d bytes",
				object.Size, written)
		}
	}

	return nil
}

func checkStoreForObjects(input io.Reader, store store.Store) ([]bdm.Object, error) {
	decompressedInput, err := util.CreateDecompressingReader(input)
	if err != nil {
		return nil, fmt.Errorf("error creating decompressing reader: %w", err)
	}
	defer decompressedInput.Close()

	objects, err := bdm.ReadObjectsFromStream(decompressedInput)
	if err != nil {
		return nil, fmt.Errorf("error reading objects from stream: %w", err)
	}

	foundObjects := make([]bdm.Object, 0)
	for _, object := range objects {
		foundObject, err := store.GetObject(object.Hash)
		if err == nil &&
			foundObject.Size == object.Size &&
			foundObject.Hash == object.Hash {
			foundObjects = append(foundObjects, *foundObject)
		}
	}

	return foundObjects, nil
}
