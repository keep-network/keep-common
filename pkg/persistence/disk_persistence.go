package persistence

import (
	"fmt"
	"io/ioutil"
	"os"
)

const (
	currentDir = "current"
	archiveDir = "archive"

	maxFileNameLength = 128
)

// NewDiskHandle creates on-disk data persistence handle
func NewDiskHandle(path string) (Handle, error) {
	err := checkStoragePermission(path)
	if err != nil {
		return nil, err
	}

	err = createDir(path, currentDir)
	if err != nil {
		return nil, err
	}

	err = createDir(path, archiveDir)
	if err != nil {
		return nil, err
	}

	return &diskPersistence{
		dataDir: path,
	}, nil
}

type diskPersistence struct {
	dataDir string
}

func (ds *diskPersistence) Save(data []byte, dirName, fileName string) error {
	if len(dirName) > maxFileNameLength {
		return fmt.Errorf(
			"the maximum directory name length of [%v] exceeded for [%v]",
			maxFileNameLength,
			dirName,
		)
	}

	if len(fileName) > maxFileNameLength {
		return fmt.Errorf(
			"the maximum file name length of [%v] exceeded for [%v]",
			maxFileNameLength,
			fileName,
		)
	}

	dirPath := ds.getStorageCurrentDirPath()
	err := createDir(dirPath, dirName)
	if err != nil {
		return err
	}

	return write(fmt.Sprintf("%s/%s/%s", dirPath, dirName, fileName), data)
}

func (ds *diskPersistence) ReadAll() (<-chan DataDescriptor, <-chan error) {
	return readAll(ds.getStorageCurrentDirPath())
}

func (ds *diskPersistence) Archive(directory string) error {
	if len(directory) > maxFileNameLength {
		return fmt.Errorf(
			"the maximum directory name length of [%v] exceeded for [%v]",
			maxFileNameLength,
			directory,
		)
	}

	from := fmt.Sprintf("%s/%s/%s", ds.dataDir, currentDir, directory)
	to := fmt.Sprintf("%s/%s/%s", ds.dataDir, archiveDir, directory)

	return moveAll(from, to)
}

func (ds *diskPersistence) getStorageCurrentDirPath() string {
	return fmt.Sprintf("%s/%s", ds.dataDir, currentDir)
}

func checkStoragePermission(dirBasePath string) error {
	_, err := ioutil.ReadDir(dirBasePath)
	if err != nil {
		return fmt.Errorf("cannot read from the storage directory: [%v]", err)
	}

	tempFile, err := ioutil.TempFile(dirBasePath, "write-test.*.tmp")
	if err != nil {
		return fmt.Errorf("cannot write to the storage directory: [%v]", err)
	}

	defer os.RemoveAll(tempFile.Name())

	return nil
}

func createDir(dirBasePath, newDirName string) error {
	dirPath := fmt.Sprintf("%s/%s", dirBasePath, newDirName)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err = os.Mkdir(dirPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error occurred while creating a dir: [%v]", err)
		}
	}

	return nil
}

// create and write data to a file
func write(filePath string, data []byte) error {
	writeFile, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer closeFile(writeFile)

	_, err = writeFile.Write(data)
	if err != nil {
		return err
	}

	err = writeFile.Sync()
	if err != nil {
		return err
	}

	return nil
}

// read a file from a file system
func read(filePath string) ([]byte, error) {
	readFile, err := os.Open(filePath) // #nosec
	if err != nil {
		return nil, err
	}

	defer closeFile(readFile)

	data, err := ioutil.ReadAll(readFile)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func closeFile(file *os.File) {
	err := file.Close()
	if err != nil {
		logger.Errorf("could not close file [%v]: [%v]", file.Name(), err)
	}
}

// readAll reads all files from the provided directoryPath and outputs them
// as DataDescriptors into the first returned output channel. All errors
// occurred during file system reading are sent to the second output channel
// returned from this function. The output can be later processed using
// pipeline pattern. This function is non-blocking and returned channels are
// not buffered. Channels are closed when there is no more to be read.
func readAll(directoryPath string) (<-chan DataDescriptor, <-chan error) {
	dataChannel := make(chan DataDescriptor)
	errorChannel := make(chan error)

	go func() {
		defer close(dataChannel)
		defer close(errorChannel)

		files, err := ioutil.ReadDir(directoryPath)
		if err != nil {
			errorChannel <- fmt.Errorf(
				"could not read the directory [%v]: [%v]",
				directoryPath,
				err,
			)
		}

		for _, file := range files {
			if file.IsDir() {
				dir, err := ioutil.ReadDir(fmt.Sprintf("%s/%s", directoryPath, file.Name()))
				if err != nil {
					errorChannel <- fmt.Errorf(
						"could not read the directory [%s/%s]: [%v]",
						directoryPath,
						file.Name(),
						err,
					)
				}

				for _, dirFile := range dir {
					// capture shared loop variables for the closure
					dirName := file.Name()
					fileName := dirFile.Name()

					readFunc := func() ([]byte, error) {
						return read(fmt.Sprintf(
							"%s/%s/%s",
							directoryPath,
							dirName,
							fileName,
						))
					}
					dataChannel <- &dataDescriptor{fileName, dirName, readFunc}
				}
			}
		}
	}()

	return dataChannel, errorChannel
}

func moveAll(directoryFromPath, directoryToPath string) error {
	_, err := os.Stat(directoryToPath)

	// target directory does not exist, we can move everything
	if os.IsNotExist(err) {
		err := os.Rename(directoryFromPath, directoryToPath)
		if err != nil {
			return fmt.Errorf("error occurred while moving a dir: [%v]", err)
		}

		return nil
	}

	// unexpected error occurred while checking target directory existence,
	// returning
	if err != nil {
		return fmt.Errorf("could not stat target directory: [%v]", err)
	}

	// target directory does exit, we need to append files
	files, err := ioutil.ReadDir(directoryFromPath)
	if err != nil {
		return fmt.Errorf("could not read directory [%v]: [%v]", directoryFromPath, err)
	}
	for _, file := range files {
		from := fmt.Sprintf("%s/%s", directoryFromPath, file.Name())
		to := fmt.Sprintf("%s/%s", directoryToPath, file.Name())
		err := os.Rename(from, to)
		if err != nil {
			return err
		}
	}
	err = os.RemoveAll(directoryFromPath)
	if err != nil {
		return fmt.Errorf("error occurred while removing archived dir: [%v]", err)
	}

	return nil
}
