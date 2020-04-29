package dal

import "time"

type FileInfo struct {
	Name               string
	DateOfCreation     time.Time
	DateOfModification time.Time
}
