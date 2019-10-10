package srv

import (
	"context"
	"sync"
	"time"

	"github.com/Tahler/isotope/convert/pkg/graph/script"
	"github.com/Tahler/isotope/service/srv/prometheus"
	grpc "google.golang.org/grpc"
	"istio.io/fortio/log"
)

// Ping checks the service graph to call its dependencies, and waits for their responses.
// It also records the execution duration.
func (s *Server) Ping(c context.Context, in *PingMessage) (*PingMessage, error) {
	log.Infof("GRPC request received!!")
	startTime := time.Now()
	prometheus.RecordRequestReceived()

	var wg sync.WaitGroup

	for _, service := range s.graph.Services {
		if service.Name == s.name {
			for _, cmd := range service.Script {

				switch requestType := cmd.(type) {
				case script.RequestCommand:
					wg.Add(1)
					go func() {
						defer wg.Done()
						s.ping(requestType.ServiceName + ":" + s.grpcPort)
					}()

				case script.ConcurrentCommand:
					numSubCmds := len(requestType)
					wg.Add(numSubCmds)
					for _, subCmd := range requestType {
						go func(step interface{}) {
							defer wg.Done()
							sc := step.(script.RequestCommand)
							s.ping(sc.ServiceName + ":" + s.grpcPort)
						}(subCmd)
					}

				case script.SleepCommand:
					time.Sleep(time.Duration(requestType))

				default:
					log.Fatalf("unknown command type in script: %T", cmd)
				}
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
// Ping returns the input ping message as an output, although we don't care about it.
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
