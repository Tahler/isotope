package main

import (
	"context"
	"log"
	"os"
	"time"

	pb "github.com/Tahler/isotope/service/srv"
	"google.golang.org/grpc"
)

const (
	defaultAddress = "127.0.0.1:8081"
	defaultMessage = "dummy"
)

func main() {
	address := defaultAddress
	if len(os.Args) > 1 {
		address = os.Args[1]
	}

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewPingServerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.Ping(ctx, &pb.PingMessage{Payload: defaultMessage})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Pinging: %s", r.Payload)
}
