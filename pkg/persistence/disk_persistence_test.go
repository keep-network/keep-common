package persistence

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
)

var (
	dataDir = "./"

	dirCurrent  = "current"
	dirArchive  = "archive"
	dirSnapshot = "snapshot"

	dirName1   = "0x424242"
	fileName11 = "file11"
	fileName12 = "file12"

	dirName2   = "0x777777"
	fileName21 = "file21"

	pathToCurrent  = fmt.Sprintf("%s/%s", dataDir, dirCurrent)
	pathToArchive  = fmt.Sprintf("%s/%s", dataDir, dirArchive)
	pathToSnapshot = fmt.Sprintf("%s/%s", dataDir, dirSnapshot)

	errExpectedRead  = fmt.Errorf("cannot read from the storage directory: ")
	errExpectedWrite = fmt.Errorf("cannot write to the storage directory: ")

	// 128 characters
	maxAllowedName = "0cc2abf49e067b3bede8426d9369c6952655d33629f44610536445c00c974c7e2740dfac967dbfceeeec0af88cd48a5d8c1c167df93cad1b8301a4a204c9f235"
	// 129 charactes
	notAllowedName = "0cc2abf49e067b3bede8426d9369c6952655d33629f44610536445c00c974c7e2740dfac967dbfceeeec0af88cd48a5d8c1c167df93cad1b8301a4a204c9f235a"

	errDirectoryNameLength = fmt.Errorf("the maximum directory name length of [128] exceeded for [%v]", notAllowedName)
	errFileNameLength      = fmt.Errorf("the maximum file name length of [128] exceeded for [%v]", notAllowedName)
)

func cleanup() {
	os.RemoveAll(pathToCurrent)
	os.RemoveAll(pathToArchive)
	os.RemoveAll(pathToSnapshot)
}

func TestDiskPersistence_Save(t *testing.T) {
	diskPersistence, _ := NewDiskHandle(dataDir)
	bytesToTest := []byte{115, 111, 109, 101, 10}

	err := diskPersistence.Save(bytesToTest, dirName1, fileName11)
	if err != nil {
		t.Fatal(err)
	}

	pathToFile := fmt.Sprintf("%s/%s/%s", pathToCurrent, dirName1, fileName11)

	if _, err := os.Stat(pathToFile); os.IsNotExist(err) {
		t.Fatalf("file [%+v] was supposed to be created", pathToFile)
	}

	cleanup()
}

func TestDiskPersistence_SaveMaxAllowed(t *testing.T) {
	diskPersistence, _ := NewDiskHandle(dataDir)
	bytesToTest := []byte{115, 111, 109, 101, 10}

	err := diskPersistence.Save(bytesToTest, maxAllowedName, maxAllowedName)
	if err != nil {
		t.Fatal(err)
	}

	pathToFile := fmt.Sprintf("%s/%s/%s", pathToCurrent, maxAllowedName, maxAllowedName)

	if _, err := os.Stat(pathToFile); os.IsNotExist(err) {
		t.Fatalf("file [%+v] was supposed to be created", pathToFile)
	}

	cleanup()
}

func TestDiskPersistence_RefuseSave(t *testing.T) {
	diskPersistence, _ := NewDiskHandle(dataDir)
	bytesToTest := []byte{115, 111, 109, 101, 10}

	err := diskPersistence.Save(bytesToTest, notAllowedName, fileName11)
	if err == nil {
		t.Fatalf("expected error")
	}
	if errDirectoryNameLength.Error() != err.Error() {
		t.Fatalf(
			"unexpected error returned\nexpected: [%v]\nactual:   [%v]",
			errDirectoryNameLength.Error(),
			err.Error(),
		)
	}

	err = diskPersistence.Save(bytesToTest, dirName1, notAllowedName)
	if err == nil {
		t.Fatalf("expected error")
	}
	if errFileNameLength.Error() != err.Error() {
		t.Fatalf(
			"unexpected error returned\nexpected: [%v]\nactual:   [%v]",
			errFileNameLength.Error(),
			err.Error(),
		)
	}

	cleanup()
}

func TestDiskPersistence_Snapshot(t *testing.T) {
	diskHandle, _ := NewDiskHandle(dataDir)
	bytesToTest := []byte{115, 111, 109, 101, 10}

	counter := 0
	diskHandle.(*diskPersistence).snapshotSuffixGenerator = func() string {
		counter++
		return fmt.Sprintf(".%d", counter)
	}

	for i := 0; i < 3; i++ {
		err := diskHandle.Snapshot(bytesToTest, dirName1, fileName11)
		if err != nil {
			t.Fatal(err)
		}
	}

	pathToFile := fmt.Sprintf(
		"%s/%s/%s",
		pathToSnapshot,
		dirName1,
		fileName11+".1",
	)

	if _, err := os.Stat(pathToFile); os.IsNotExist(err) {
		t.Fatalf("file [%+v] was supposed to be created", pathToFile)
	}

	pathToFile = fmt.Sprintf(
		"%s/%s/%s",
		pathToSnapshot,
		dirName1,
		fileName11+".2",
	)

	if _, err := os.Stat(pathToFile); os.IsNotExist(err) {
		t.Fatalf("file [%+v] was supposed to be created", pathToFile)
	}

	pathToFile = fmt.Sprintf(
		"%s/%s/%s",
		pathToSnapshot,
		dirName1,
		fileName11+".3",
	)

	if _, err := os.Stat(pathToFile); os.IsNotExist(err) {
		t.Fatalf("file [%+v] was supposed to be created", pathToFile)
	}

	cleanup()
}

func TestDiskPersistence_SnapshotMaxAllowed(t *testing.T) {
	diskHandle, _ := NewDiskHandle(dataDir)
	bytesToTest := []byte{115, 111, 109, 101, 10}

	snapshotSuffix := ".suffix"
	diskHandle.(*diskPersistence).snapshotSuffixGenerator = func() string {
		return snapshotSuffix
	}

	maxAllowedSnapshotName := maxAllowedName[0 : len(maxAllowedName)-len(snapshotSuffix)]
	err := diskHandle.Snapshot(bytesToTest, maxAllowedName, maxAllowedSnapshotName)
	if err != nil {
		t.Fatal(err)
	}

	pathToFile := fmt.Sprintf(
		"%s/%s/%s",
		pathToSnapshot,
		maxAllowedName,
		maxAllowedSnapshotName+snapshotSuffix,
	)

	if _, err := os.Stat(pathToFile); os.IsNotExist(err) {
		t.Fatalf("file [%+v] was supposed to be created", pathToFile)
	}

	cleanup()
}

func TestDiskPersistence_RefuseSnapshot_MaxAllowedExceeded(t *testing.T) {
	diskHandle, _ := NewDiskHandle(dataDir)
	bytesToTest := []byte{115, 111, 109, 101, 10}

	snapshotSuffix := ".suffix"
	diskHandle.(*diskPersistence).snapshotSuffixGenerator = func() string {
		return snapshotSuffix
	}

	err := diskHandle.Snapshot(bytesToTest, notAllowedName, fileName11)
	if err == nil {
		t.Fatalf("expected error")
	}

	if errDirectoryNameLength.Error() != err.Error() {
		t.Fatalf(
			"unexpected error returned\nexpected: [%v]\nactual:   [%v]",
			errDirectoryNameLength.Error(),
			err.Error(),
		)
	}

	err = diskHandle.Snapshot(bytesToTest, dirName1, notAllowedName)
	if err == nil {
		t.Fatalf("expected error")
	}

	errSnapshotFileNameLength := fmt.Errorf(
		"the maximum file name length of [%d] exceeded for [%v]",
		128-len(snapshotSuffix),
		notAllowedName,
	)

	if errSnapshotFileNameLength.Error() != err.Error() {
		t.Fatalf(
			"unexpected error returned\nexpected: [%v]\nactual:   [%v]",
			errFileNameLength.Error(),
			err.Error(),
		)
	}

	cleanup()
}

func TestDiskPersistence_RefuseSnapshot_NameCollision(t *testing.T) {
	diskHandle, _ := NewDiskHandle(dataDir)
	bytesToTest := []byte{115, 111, 109, 101, 10}

	// snapshot suffix generator return always the same suffix in order to
	// cause name collision.
	snapshotSuffix := ".suffix"
	diskHandle.(*diskPersistence).snapshotSuffixGenerator = func() string {
		return snapshotSuffix
	}

	err := diskHandle.Snapshot(bytesToTest, dirName1, fileName11)
	if err != nil {
		t.Fatal(err)
	}

	pathToFile := fmt.Sprintf(
		"%s/%s/%s",
		pathToSnapshot,
		dirName1,
		fileName11+snapshotSuffix,
	)

	if _, err := os.Stat(pathToFile); os.IsNotExist(err) {
		t.Fatalf("file [%+v] was supposed to be created", pathToFile)
	}

	err = diskHandle.Snapshot(bytesToTest, dirName1, fileName11)

	expectedDuplicateError := fmt.Errorf(
		"could not create unique snapshot; " +
			"snapshot name collision has been detected",
	)
	if err == nil || (!reflect.DeepEqual(expectedDuplicateError, err)) {
		t.Fatalf(
			"unexpected error\nexpected: [%v]\nactual:   [%v]",
			expectedDuplicateError,
			err,
		)
	}

	cleanup()
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

	cleanup()
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

	cleanup()
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

	cleanup()
}

func TestDiskPersistence_ArchiveMaxAllowed(t *testing.T) {
	diskPersistence, _ := NewDiskHandle(dataDir)

	pathMoveFrom := fmt.Sprintf("%s/%s", pathToCurrent, maxAllowedName)
	pathMoveTo := fmt.Sprintf("%s/%s", pathToArchive, maxAllowedName)

	bytesToTest := []byte{115, 111, 109, 101, 10}

	diskPersistence.Save(bytesToTest, maxAllowedName, maxAllowedName)

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

	err := diskPersistence.Archive(maxAllowedName)
	if err != nil {
		t.Fatal(err)
	}

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

	cleanup()
}

func TestDiskPersistence_RefuseArchive(t *testing.T) {
	diskPersistence, _ := NewDiskHandle(dataDir)

	err := diskPersistence.Archive(notAllowedName)
	if err == nil {
		t.Fatalf("expected error")
	}
	if errDirectoryNameLength.Error() != err.Error() {
		t.Fatalf(
			"unexpected error returned\nexpected: [%v]\nactual:   [%v]",
			errDirectoryNameLength.Error(),
			err.Error(),
		)
	}

	cleanup()
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

	cleanup()
}
