package bdm

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/cry-inc/bdm/pkg/bdm/util"
)

// 10 MB should be enough for the objects JSON data!
const sizeLimit = 10 * 1024 * 1024

// ReadObjectsFromStream reads a list of Objects from a stream
func ReadObjectsFromStream(reader io.Reader) ([]Object, error) {
	lengthData := make([]byte, 8)
	_, err := io.ReadFull(reader, lengthData)
	if err != nil {
		return nil, fmt.Errorf("error reading objects length: %w", err)
	}

	length, err := util.Int64FromBytes(lengthData)
	if err != nil {
		return nil, fmt.Errorf("error decoding objects length: %w", err)
	}
	if length <= 0 || length >= sizeLimit {
		return nil, fmt.Errorf("found invalid JSON length %d", length)
	}

	jsonData := make([]byte, length)
	_, err = io.ReadFull(reader, jsonData)
	if err != nil {
		return nil, fmt.Errorf("error reading objects JSON data: %w", err)
	}

	var objects []Object
	err = json.Unmarshal(jsonData, &objects)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling objects JSON: %w", err)
	}

	return objects, nil
}

// WriteObjectsToStream writes a list of Objects to a stream
func WriteObjectsToStream(objects []Object, output io.Writer) error {
	jsonData, err := json.Marshal(objects)
	if err != nil {
		return fmt.Errorf("error marshalling objects JSON: %w", err)
	}

	length := int64(len(jsonData))
	lengthData := util.Int64ToBytes(length)
	written, err := output.Write(lengthData)
	if err != nil {
		return fmt.Errorf("error writing objects JSON length: %w", err)
	}
	if written != len(lengthData) {
		return fmt.Errorf("error writing objects JSON length: wrote %d of %d bytes",
			written, len(lengthData))
	}

	written, err = output.Write(jsonData)
	if err != nil {
		return fmt.Errorf("error writing objects JSON: %w", err)
	}
	if written != len(jsonData) {
		return fmt.Errorf("error writing objects JSON: wrote %d of %d bytes",
			written, len(jsonData))
	}

	return nil
}
