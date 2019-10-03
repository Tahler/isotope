package srv

import (
	context "context"
	"fmt"
	"time"

	grpc "google.golang.org/grpc"
	"istio.io/fortio/log"
)

// GRPC ping
func (s *Server) Ping(c context.Context, in *PingMessage) (*PingMessage, error) {
	fmt.Println("Request received!!")
	return in, nil
}

func ping(ch chan struct{}, address string) {

	log.Infof("Pinging to: %v", address)
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := NewPingServerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for {
		select {
		case msg := <-ch:
			r, err := c.Ping(ctx, &PingMessage{})
			if err != nil {
				log.Fatalf("could not ping: %v", err)
			}
			fmt.Println("#####", c, r, msg)
		}
	}
}
