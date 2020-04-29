package main

import (
	"context"
	files_api "github.com/go-project/proto/generated"
	"github.com/rs/xid"
	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"io"
	"io/ioutil"
	"testing"
	"time"
)

func TestUploadFileSvc(t *testing.T) {
	go main()

	Convey("Connect to grpc and create client", t, func() {
		conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
		defer conn.Close()
		So(err, ShouldBeNil)

		client := files_api.NewFilesSvcClient(conn)

		Convey("Prepare data", func() {

			bytes, err := ioutil.ReadFile("Original-3.jpg")
			So(err, ShouldBeNil)

			fileHeader := &files_api.File_FileHeader{
				FileHeader: &files_api.FileHeader{
					Name: xid.New().String() + ".jpg",
					Size: int64(len(bytes)),
				},
			}

			stream, err := client.Upload(context.Background())
			So(err, ShouldBeNil)

			Convey("Send hole file", func() {

				err := stream.Send(&files_api.File{FilePart: fileHeader})
				So(err, ShouldBeNil)

				err = stream.Send(&files_api.File{FilePart: &files_api.File_FileChunk{FileChunk: &files_api.FileChunk{Data: bytes}}})
				So(err, ShouldBeNil)

				_, err = stream.CloseAndRecv()
				So(err, ShouldBeNil)

			})

			Convey("Send part file (for big size)", func() {

				err := stream.Send(&files_api.File{FilePart: fileHeader})
				So(err, ShouldBeNil)

				middle := len(bytes) / 2

				err = stream.Send(&files_api.File{FilePart: &files_api.File_FileChunk{FileChunk: &files_api.FileChunk{Data: bytes[:middle]}}})
				So(err, ShouldBeNil)

				err = stream.Send(&files_api.File{FilePart: &files_api.File_FileChunk{FileChunk: &files_api.FileChunk{Data: bytes[middle:]}}})
				So(err, ShouldBeNil)

				_, err = stream.CloseAndRecv()
				So(err, ShouldBeNil)

			})

			Convey("Send part file not complete (for big size)", func() {

				err := stream.Send(&files_api.File{FilePart: fileHeader})
				So(err, ShouldBeNil)

				middle := len(bytes) / 2

				err = stream.Send(&files_api.File{FilePart: &files_api.File_FileChunk{FileChunk: &files_api.FileChunk{Data: bytes[:middle]}}})
				So(err, ShouldBeNil)

				_, err = stream.CloseAndRecv()
				So(err.Error(), ShouldEqual, "rpc error: code = Canceled desc = file recording not complete")

			})

			Convey("Send chunk part only", func() {

				err := stream.Send(&files_api.File{FilePart: &files_api.File_FileChunk{FileChunk: &files_api.FileChunk{Data: bytes}}})
				So(err, ShouldBeNil)

				_, err = stream.CloseAndRecv()
				So(err.Error(), ShouldEqual, "rpc error: code = InvalidArgument desc = file chunk before header")

			})
		})
	})
}

func TestGetListSvc(t *testing.T) {
	go main()

	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	defer conn.Close()

	if err != nil {
		t.Error(err)
	}

	getList := func(numb int) {

		client := files_api.NewFilesSvcClient(conn)

		stream, err := client.GetList(context.Background(), &files_api.FilesGetListReq{})
		if err != nil {
			t.Error(err)
		}

		for {
			fileInfo, err := stream.Recv()
			if err == io.EOF {
				t.Log("eof")
				break
			}
			if err != nil {
				t.Error(err)
				break
			}
			t.Log(fileInfo)
		}
	}

	for i := 0; i < 220; i++ {
		go getList(i)
	}

	time.Sleep(time.Minute * 1)
}
