package main

import (
	"bufio"
	proto "chat/proto"
	"crypto/sha256"
	"flag"
	"fmt"
	"os"

	"encoding/hex"
	"log"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var client proto.BroadcastClient
var wait *sync.WaitGroup

func init() { // Initialize the wait group
	wait = &sync.WaitGroup{}
}

func connect(user *proto.User) error {
	var streamerror error

	stream, err := client.CreateStream(context.Background(), &proto.Connect{
		User:   user,
		Active: true,
	})

	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}

	wait.Add(1)
	go func(str proto.Broadcast_CreateStreamClient) {
		defer wait.Done()

		for {
			msg, err := str.Recv() // Wait for us to receive a message from the server
			if err != nil {
				streamerror = fmt.Errorf("Error reading message: %v", err)
				break
			}

			fmt.Printf("%v : %s\n", msg.Id, msg.Content) // Print the message that you received from the server. So you can see in the terminal

		}
	}(stream)

	return streamerror
}

func main() {
	timestamp := time.Now()
	done := make(chan int)

	name := flag.String("N", "Anon", "The name of the user") // User can enter their username - N is the flag. The default name is Anon.  Description of the field.
	flag.Parse()

	id := sha256.Sum256([]byte(timestamp.String() + *name)) // Generate id by using the time stamp and hash it.

	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure()) // Connect to the local host. No https so it is insecure.
	if err != nil {
		log.Fatalf("Couldnt connect to service: %v", err)
	}

	client = proto.NewBroadcastClient(conn)
	user := &proto.User{ // Create the user
		Id:   hex.EncodeToString(id[:]),
		Name: *name,
	}

	connect(user)

	wait.Add(1) // Why do we use a wait group here?
	go func() {
		defer wait.Done()

		scanner := bufio.NewScanner(os.Stdin) // Scan the message input from the user
		for scanner.Scan() {
			msg := &proto.Message{
				Id:        user.Id,
				Content:   scanner.Text(),
				Timestamp: timestamp.String(),
			}

			_, err := client.BroadcastMessage(context.Background(), msg) // Send the message to the server?
			if err != nil {
				fmt.Printf("Error Sending Message: %v", err)
				break
			}
		}

	}()

	go func() {
		wait.Wait()
		close(done)
	}()

	<-done
}
