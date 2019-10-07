package srv

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Tahler/isotope/convert/pkg/graph/script"
	"github.com/Tahler/isotope/service/srv/prometheus"
	grpc "google.golang.org/grpc"
	"istio.io/fortio/log"
)

// Ping checks the service graph to call its dependencies, waits for their responses,
// and returns the input ping message as an output.
// It also records the execution duration.
func (s *Server) Ping(c context.Context, in *PingMessage) (*PingMessage, error) {
	fmt.Println("GRPC request received!!")
	startTime := time.Now()
	prometheus.RecordRequestReceived()

	var wg sync.WaitGroup

	for _, service := range s.graph.Services {
		// fmt.Println(s.name, service.Name)
		if service.Name == s.name {
			for _, cmd := range service.Script {
				fmt.Printf("SCRIPT %+v", cmd)
				c := cmd.(script.RequestCommand)
				wg.Add(1)
				go func() {
					defer wg.Done()
					// TODO: Include more types of commands: sleep,...
					s.ping(c.ServiceName + ":" + s.grpcPort)
				}()
			}
		}
	}

	wg.Wait()

	stopTime := time.Now()
	duration := stopTime.Sub(startTime)
	prometheus.RecordResponseSent(duration, 0, 200)

	return in, nil
}

// ping method starts a grpc client and make ping to the destination address.
func (s *Server) ping(address string) {

	log.Infof("Pinging to: %v", address)
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := NewPingServerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.Ping(ctx, &PingMessage{})
	if err != nil {
		log.Infof("could not ping: %v", err)
	}
}
