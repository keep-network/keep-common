package persistence

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"
)

var (
	dataDir = "./"

	dirCurrent = "current"
	dirArchive = "archive"

	dirName1   = "0x424242"
	fileName11 = "/file11"
	fileName12 = "/file12"

	dirName2   = "0x777777"
	fileName21 = "/file21"

	pathToCurrent = fmt.Sprintf("%s/%s", dataDir, dirCurrent)
	pathToArchive = fmt.Sprintf("%s/%s", dataDir, dirArchive)

	errExpectedRead  = errors.New("cannot read from the storage directory: ")
	errExpectedWrite = errors.New("cannot write to the storage directory: ")
)

func TestMain(m *testing.M) {
	code := m.Run()
	os.RemoveAll(pathToCurrent)
	os.RemoveAll(pathToArchive)
	os.Exit(code)
}

func TestDiskPersistence_Save(t *testing.T) {
	diskPersistence, _ := NewDiskHandle(dataDir)
	bytesToTest := []byte{115, 111, 109, 101, 10}

	diskPersistence.Save(bytesToTest, dirName1, fileName11)

	pathToFile := fmt.Sprintf("%s/%s%s", pathToCurrent, dirName1, fileName11)

	if _, err := os.Stat(pathToFile); os.IsNotExist(err) {
		t.Fatalf("file [%+v] was supposed to be created", pathToFile)
	}
}

func TestDiskPersistence_StoragePermission(t *testing.T) {
	tempDir := "./data_storage"

	err := os.Mkdir(tempDir, 0000) // d---------
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Fatalf("dir [%+v] was supposed to be created", tempDir)
	}
	defer os.RemoveAll(tempDir)

	_, err = NewDiskHandle(tempDir)
	if err == nil || !strings.Contains(err.Error(), errExpectedRead.Error()) {
		t.Fatalf("error on read was supposed to be returned")
	}

	os.Chmod(tempDir, 0444) // dr--r--r

	_, err = NewDiskHandle(tempDir)
	if err == nil || !strings.Contains(err.Error(), errExpectedWrite.Error()) {
		t.Fatalf("error on write was supposed to be returned")
	}
}

func TestDiskPersistence_ReadAll(t *testing.T) {
	diskPersistence, _ := NewDiskHandle(dataDir)

	bytesToTest := []byte{115, 111, 109, 101, 10}
	expectedBytes := [][]byte{bytesToTest, bytesToTest, bytesToTest}

	diskPersistence.Save(bytesToTest, dirName1, fileName11)
	diskPersistence.Save(bytesToTest, dirName1, fileName12)
	diskPersistence.Save(bytesToTest, dirName2, fileName21)

	dataChannel, errChannel := diskPersistence.ReadAll()

	var descriptors []DataDescriptor
	var errors []error

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		for e := range errChannel {
			errors = append(errors, e)
		}
		wg.Done()
	}()

	go func() {
		for d := range dataChannel {
			descriptors = append(descriptors, d)
		}
		wg.Done()
	}()

	wg.Wait()

	for err := range errors {
		t.Fatal(err)
	}

	if len(descriptors) != 3 {
		t.Fatalf(
			"Number of descriptors does not match\nExpected: [%v]\nActual:   [%v]",
			3,
			len(descriptors),
		)
	}

	for i := 0; i < 3; i++ {
		fileContent, err := descriptors[i].Content()
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(expectedBytes[i], fileContent) {
			t.Errorf(
				"unexpected file content [%d]\nexpected: [%v]\nactual:   [%v]\n",
				i,
				expectedBytes[i],
				fileContent,
			)
		}
	}
}

func TestDiskPersistence_Archive(t *testing.T) {
	diskPersistence, _ := NewDiskHandle(dataDir)

	pathMoveFrom := fmt.Sprintf("%s/%s", pathToCurrent, dirName1)
	pathMoveTo := fmt.Sprintf("%s/%s", pathToArchive, dirName1)

	bytesToTest := []byte{115, 111, 109, 101, 10}

	diskPersistence.Save(bytesToTest, dirName1, fileName11)

	if _, err := os.Stat(pathMoveFrom); os.IsNotExist(err) {
		if err != nil {
			t.Fatalf("Dir [%+v] was supposed to be created", pathMoveFrom)
		}
	}

	if _, err := os.Stat(pathMoveTo); !os.IsNotExist(err) {
		if err != nil {
			t.Fatalf("Dir [%+v] was supposed to be empty", pathMoveTo)
		}
	}

	diskPersistence.Archive(dirName1)

	if _, err := os.Stat(pathMoveFrom); !os.IsNotExist(err) {
		if err != nil {
			t.Fatalf("Dir [%+v] was supposed to be moved", pathMoveFrom)
		}
	}

	if _, err := os.Stat(pathMoveTo); os.IsNotExist(err) {
		if err != nil {
			t.Fatalf("Dir [%+v] was supposed to be created", pathMoveTo)
		}
	}
}

func TestDiskPersistence_AppendToArchive(t *testing.T) {
	diskPersistence, _ := NewDiskHandle(dataDir)

	pathMoveFrom := fmt.Sprintf("%s/%s", pathToCurrent, dirName1)
	pathMoveTo := fmt.Sprintf("%s/%s", pathToArchive, dirName1)

	bytesToTest := []byte{115, 111, 109, 101, 10}

	diskPersistence.Save(bytesToTest, dirName1, fileName11)
	diskPersistence.Save(bytesToTest, dirName1, fileName12)
	diskPersistence.Archive(dirName1)

	diskPersistence.Save(bytesToTest, dirName1, "/file13")
	diskPersistence.Save(bytesToTest, dirName1, "/file14")
	diskPersistence.Archive(dirName1)

	if _, err := os.Stat(pathMoveFrom); !os.IsNotExist(err) {
		t.Fatalf("Dir [%+v] was supposed to be removed", pathMoveFrom)
	}

	files, _ := ioutil.ReadDir(pathMoveTo)
	if len(files) != 4 {
		t.Fatalf("Number of all files was supposed to be [%+v]", 4)
	}

}
