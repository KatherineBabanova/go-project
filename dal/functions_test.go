package dal

import (
	"io/ioutil"
	"testing"
)

func TestFile(t *testing.T) {

	bytes, err := ioutil.ReadFile("Original-3.jpg")
	if err != nil {
		t.Error("file open error", err)
	}

	osFile, err := CreateFile("", "newFile.jpg")
	if err != nil {
		t.Error("file create error", err)
	}

	err = AppendFile(osFile, bytes[:1093015])
	if err != nil {
		t.Error("file append error", err)
	}

	//err = AppendFile(osFile, bytes[1093015:])
	//if err != nil {
	//	t.Log("file append2 error", err)
	//}
}

func TestList(t *testing.T) {
	filesInfo, err := GetFilesList("./")
	if err != nil {
		t.Error("GetFilesList error", err)
	}
	for _, fileInfo := range filesInfo {
		t.Log(fileInfo.Name, fileInfo.DateOfCreation, fileInfo.DateOfModification)
	}
}
