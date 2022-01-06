package store

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/cry-inc/bdm/pkg/bdm"
	"github.com/cry-inc/bdm/pkg/bdm/util"
)

const sizeSuffix = "_size"

func getObjectPath(hash string) string {
	if len(hash) <= 2 {
		return hash
	}

	folder := hash[0:2]
	file := hash[2:]
	return path.Join(folder, file)
}

func (s *packageStore) GetObject(hash string) (*bdm.Object, error) {
	objectPath := getObjectPath(hash)
	filePath := path.Join(s.objectsFolder, objectPath)
	if !util.FileExists(filePath) {
		return nil, fmt.Errorf("unable to find object file %s", filePath)
	}

	sizePath := filePath + sizeSuffix
	sizeBytes, err := os.ReadFile(sizePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read object size file %s: %w",
			sizePath, err)
	}

	size, err := util.Int64FromBytes(sizeBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing object size: %w", err)
	}

	return &bdm.Object{
		Hash: hash,
		Size: size,
	}, nil
}

func (s *packageStore) AddObject(reader io.Reader) (*bdm.Object, error) {
	tempFile, err := os.CreateTemp(s.objectsFolder, "tmp_*")
	if err != nil {
		return nil, fmt.Errorf("error opening temporary object file: %w", err)
	}
	defer tempFile.Close()

	compressedHandle, err := util.CreateCompressingWriter(tempFile)
	if err != nil {
		return nil, fmt.Errorf("error creating compressing writer: %w", err)
	}
	defer compressedHandle.Close()

	hasher := util.CreateHasher()
	writer := io.MultiWriter(compressedHandle, hasher)

	fileSize, err := io.Copy(writer, reader)
	if err != nil {
		return nil, fmt.Errorf("error writing compressed object data: %w", err)
	}

	hash := util.GetHashString(hasher)
	tempFileName := tempFile.Name()

	compressedHandle.Close()
	tempFile.Close()

	{
		s.objectsMutex.Lock()
		defer s.objectsMutex.Unlock()

		object, _ := s.GetObject(hash)
		if object != nil {
			// Object exists already in store, no file moving required!
			os.Remove(tempFileName)
		} else {
			objectPath := getObjectPath(hash)
			finalPath := path.Join(s.objectsFolder, objectPath)
			finalFolder := path.Dir(finalPath)
			if !util.FolderExists(finalFolder) {
				err = os.MkdirAll(finalFolder, os.ModePerm)
				if err != nil {
					return nil, fmt.Errorf("error creating object folder %s: %w",
						finalFolder, err)
				}
			}

			err := os.Rename(tempFileName, finalPath)
			if err != nil {
				return nil, fmt.Errorf("error finalizing object file name: %w", err)
			}

			sizeBytes := util.Int64ToBytes(fileSize)
			sizePath := finalPath + sizeSuffix
			err = os.WriteFile(sizePath, sizeBytes, os.ModePerm)
			if err != nil {
				return nil, fmt.Errorf("error writing object size file %s: %w",
					sizePath, err)
			}
		}
	}

	object := bdm.Object{
		Hash: hash,
		Size: fileSize,
	}

	return &object, nil
}

func (s *packageStore) ReadObject(hash string) (io.ReadCloser, error) {
	objectPath := getObjectPath(hash)
	filePath := path.Join(s.objectsFolder, objectPath)
	fileHandle, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening object file: %w", err)
	}

	decompressedHandle, err := util.CreateDecompressingReader(fileHandle)
	if err != nil {
		return nil, fmt.Errorf("error creating decompressing reader: %w", err)
	}

	reader, writer := io.Pipe()
	go func() {
		defer fileHandle.Close()
		defer decompressedHandle.Close()
		defer writer.Close()
		_, err := io.Copy(writer, decompressedHandle)
		if err != nil {
			panic(fmt.Errorf("error reading compressed object data from file %s: %w", filePath, err))
		}
	}()

	return reader, nil
}

func (s *packageStore) GetObjects() ([]*bdm.Object, error) {
	if !util.FolderExists(s.objectsFolder) {
		return nil, fmt.Errorf("objects store folder %s does not exist",
			s.objectsFolder)
	}

	folders, err := os.ReadDir(s.objectsFolder)
	if err != nil {
		return nil, fmt.Errorf("error reading object store directory: %w", err)
	}

	objects := make([]*bdm.Object, 0)
	for _, folder := range folders {
		if !folder.IsDir() {
			// We are only interested in the folders and ignore any files
			continue
		}
		folderPath := path.Join(s.objectsFolder, folder.Name())
		files, err := os.ReadDir(folderPath)
		if err != nil {
			return nil, fmt.Errorf("error reading object store subdirectory %s: %w",
				folderPath, err)
		}
		for _, file := range files {
			if file.Type().IsRegular() {
				name := folder.Name() + file.Name()
				if !strings.HasSuffix(name, sizeSuffix) {
					object, err := s.GetObject(name)
					if err != nil {
						return nil, fmt.Errorf("error getting object %s: %w",
							name, err)
					}
					objects = append(objects, object)
				}
			}
		}
	}

	return objects, nil

}
