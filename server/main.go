package main

import (
	proto "chat/proto"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
	glog "google.golang.org/grpc/grpclog"
)

var grpcLog glog.LoggerV2

func init() {
	// Just log some data
	grpcLog = glog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout) // Info, warning, error messages
}

type Connection struct {
	stream proto.Broadcast_CreateStreamServer // Allow us to stream messages between server and client
	id     string
	active bool
	error  chan error // since we use goroutines error type should be a channel
}

type Server struct {
	proto.UnimplementedBroadcastServer
	Connection []*Connection // Collection oc connections
}

func (s *Server) CreateStream(pconn *proto.Connect, stream proto.Broadcast_CreateStreamServer) error {
	conn := &Connection{
		stream: stream,
		id:     pconn.User.Id,
		active: true,
		error:  make(chan error),
	}

	s.Connection = append(s.Connection, conn)

	fmt.Print(pconn.User.Name)

	return <-conn.error // Where does the error appear? When we append to the connection?
}

func (s *Server) BroadcastMessage(ctx context.Context, msg *proto.Message) (*proto.Close, error) {
	wait := sync.WaitGroup{} // Waits go routine to finish and decrement the counter accordingly. This way we can block until our go routines finish
	done := make(chan int)   // Use this to know when all the goroutines are finished

	for _, conn := range s.Connection {
		wait.Add(1) // There will be a go routine for the each connection

		go func(msg *proto.Message, conn *Connection) {
			defer wait.Done() // When the goroutine finished the wait group will be decremented

			if conn.active {
				err := conn.stream.Send(msg)                      // We will grab the message from the stream and send it back to the client who is attached to this connection
				grpcLog.Info("Sending message to: ", conn.stream) // What is a stream here?

				if err != nil {
					grpcLog.Errorf("Error with Stream: %v - Error: %v", conn.stream, err)
					conn.active = false
					conn.error <- err
				}
			}
		}(msg, conn)

	}

	go func() {
		wait.Wait() // Wait for all other goroutines to exit
		close(done) // When it finishes we close the done channel
	}()

	<-done // Block the return of the statement
	return &proto.Close{}, nil
}

func main() {
	// Server setup on 8000 to listen the connections?
	var connections []*Connection

	server := &Server{Connection: connections}

	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("error creating the server %v", err)
	}

	grpcLog.Info("Starting server at port :8080")

	proto.RegisterBroadcastServer(grpcServer, server)
	grpcServer.Serve(listener)
}
