package main

//go:generate protoc -I ./proto --go_out=plugins=grpc:./proto/generated proto/files.proto
