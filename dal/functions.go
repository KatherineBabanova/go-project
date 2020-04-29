package dal

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

func GetFilesList(directory string) ([]FileInfo, error) {
	if err := checkDirectory(directory); err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	var filesInfo []FileInfo

	for _, file := range files {

		dCreation, dModification, err := getFileDates(filepath.Join(directory, file.Name()))
		if err != nil {
			return nil, err
		}

		fileInfo := FileInfo{
			Name:               file.Name(),
			DateOfCreation:     dCreation,
			DateOfModification: dModification,
		}

		filesInfo = append(filesInfo, fileInfo)
	}

	return filesInfo, nil
}

func convertTimespecToTime(ts syscall.Timespec) time.Time {
	return time.Unix(int64(ts.Sec), int64(ts.Nsec))
}

func getFileDates(fileName string) (time.Time, time.Time, error) {

	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	stat := fileInfo.Sys().(*syscall.Stat_t)

	creationTime := convertTimespecToTime(stat.Ctimespec)
	modificationTime := convertTimespecToTime(stat.Mtimespec)

	return creationTime, modificationTime, nil
}

func checkDirectory(directory string) error {

	if _, err := os.Stat(directory); os.IsNotExist(err) {

		if err := os.MkdirAll(directory, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func CreateFile(directory, fileName string) (*os.File, error) {

	if err := checkDirectory(directory); err != nil {
		return nil, err
	}

	file, err := os.Create(filepath.Join(directory, fileName))
	if err != nil {
		return nil, err
	}

	return file, nil
}

func AppendFile(file *os.File, imgBytes []byte) error {

	if _, err := file.Write(imgBytes); err != nil {
		return err
	}
	return nil
}

func DeleteFile(directory, fileName string) error {

	if err := os.Remove(filepath.Join(directory, fileName)); err != nil {

		if os.IsNotExist(err) {
			return nil
		}

		return err
	}
	return nil
}
