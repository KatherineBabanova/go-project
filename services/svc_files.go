package services

import (
	"github.com/go-project/dal"
	files_api "github.com/go-project/proto/generated"
	"github.com/rs/xid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"os"
)

var _ files_api.FilesSvcServer = new(FilesSvc)

type FilesSvc struct {
	logger     *zap.Logger
	directory  string
	inputChan  chan interface{}
	outputChan chan interface{}
}

func NewFilesSvc(logger *zap.Logger) *FilesSvc {
	logger = logger.Named("FilesSvc")

	return &FilesSvc{
		logger:     logger,
		directory:  "filesDir",
		inputChan:  make(chan interface{}, 10),
		outputChan: make(chan interface{}, 100),
	}
}

func (svc *FilesSvc) Upload(server files_api.FilesSvc_UploadServer) error {

	svc.inputChan <- struct{}{}

	defer func() {
		<-svc.inputChan
	}()

	var osFile *os.File
	var fileSize int = 0
	var fileName string

	for {
		file, err := server.Recv()

		if err == io.EOF {
			if fileSize != 0 {
				if err := dal.DeleteFile(svc.directory, fileName); err != nil {
					return status.Error(codes.Unknown, "file recording not complete. "+err.Error())
				}
				return status.Error(codes.Canceled, "file recording not complete")
			}
			return server.SendAndClose(&files_api.FilesUploadResp{})
		}
		if err != nil {
			return status.Error(codes.Unknown, err.Error())
		}

		switch filePart := file.FilePart.(type) {
		case *files_api.File_FileHeader:

			if filePart.FileHeader.Size == 0 {
				return status.Error(codes.InvalidArgument, "file size not set")
			}
			if len(filePart.FileHeader.Name) == 0 {
				fileName = xid.New().String() + ".jpg"
				svc.logger.Info("File name not found. Created new ", zap.String("name", fileName))
				//or return error
			}

			osFile, err = dal.CreateFile(svc.directory, filePart.FileHeader.Name)
			if err != nil {
				return status.Error(codes.Unknown, err.Error())
			}

			fileSize = int(filePart.FileHeader.Size)
			fileName = filePart.FileHeader.Name

		case *files_api.File_FileChunk:

			if osFile == nil {
				return status.Error(codes.InvalidArgument, "file chunk before header")
			}

			if err = dal.AppendFile(osFile, filePart.FileChunk.Data); err != nil {
				return status.Error(codes.Unknown, err.Error())
			}

			fileSize -= len(filePart.FileChunk.Data)
		}
	}

}

func (svc *FilesSvc) GetList(req *files_api.FilesGetListReq, server files_api.FilesSvc_GetListServer) error {

	svc.outputChan <- struct{}{}

	defer func() {
		<-svc.outputChan
	}()

	list, err := dal.GetFilesList(svc.directory)
	if err != nil {
		return status.Error(codes.Unknown, err.Error())
	}

	for _, fileInfo := range list {
		if err = server.Send(dao2dto_FileInfo(fileInfo)); err != nil {
			return status.Error(codes.Unknown, err.Error())
		}
	}
	return nil
}

func dao2dto_FileInfo(info dal.FileInfo) *files_api.FileInfo {

	return &files_api.FileInfo{
		Name:               info.Name,
		DateOfCreation:     info.DateOfCreation.String(),
		DateOfModification: info.DateOfModification.String(),
	}
}
