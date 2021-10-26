package main

import (
	"os"

	"google.golang.org/grpc/encoding/proto"
	glog "google.golang.org/grpc/grpclog"
)

var grpclog glog.LoggerV2

func init() {
	grpclog = glog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)
}

type Connection struct {
	stream proto.Broadcast_CreateStreamServer
	id     string
	active bool
	error  chan error
}

func main() {

}
