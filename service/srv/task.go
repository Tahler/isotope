package srv

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/Tahler/isotope/convert/pkg/graph/script"
	"github.com/Tahler/isotope/convert/pkg/graph/svctype"
	"github.com/Tahler/isotope/service/srv/prometheus"
	"github.com/fortio/fortio/log"
)

type Task struct {
	tType    script.Command
	protocol svctype.ServiceType
	dest     string
	httpUrl  string
	payload  uint64
}

func newTask(command script.Command, protocol svctype.ServiceType, dest string, url string, payload uint64) *Task {
	return &Task{
		tType:    command,
		protocol: protocol,
		dest:     dest,
		httpUrl:  url,
		payload:  payload,
	}
}

func (s *Server) executeTasks(tasks []*Task) (err error) {
	var wg sync.WaitGroup

	errc := make(chan error)
	done := make(chan bool)

	for _, task := range tasks {
		switch cmd := task.tType.(type) {
		case script.RequestCommand:
			wg.Add(1)

			go func(task *Task, errc chan error) {
				defer wg.Done()
				if task.tType == svctype.ServiceGRPC {
					err = s.ping(task.dest, task.payload)
				} else {
					err = s.executeRequestCommand(task.dest, task.httpUrl, task.payload)
				}
				// Send errors to error channel.
				if err != nil {
					errc <- err
				}
			}(task, errc)

		case script.SleepCommand:
			time.Sleep(time.Duration(cmd))
		default:
			log.Fatalf("unknown command type in script: %T", cmd)
		}
	}

	wg.Wait()
	done <- true

	// Exit if error, or if all task finished successfully.
	for {
		select {
		case err := <-errc:
			return err
		case <-done:
			return nil
		}
	}

}

/*
	HTTP
*/

// Execute sends an HTTP request to another service. Assumes DNS is available
// which maps exe.ServiceName to the relevant URL to reach the service.
func (s *Server) executeRequestCommand(destination string, url string, payload uint64) (err error) {

	response, err := s.sendRequest(destination, url, payload)
	if err != nil {
		return
	}

	prometheus.RecordRequestSent(destination, payload)

	if response.StatusCode == http.StatusInternalServerError {
		err = fmt.Errorf("service %s responded with %s", destination, response.Status)
	}

	// Necessary for reusing HTTP/1.x "keep-alive" TCP connections.
	// https://golang.org/pkg/net/http/#Response
	_, _ = io.Copy(ioutil.Discard, response.Body)
	response.Body.Close()

	return
}

func (s *Server) sendRequest(address string, url string, payloadSize uint64) (*http.Response, error) {
	// Build request
	payload := make([]byte, payloadSize)
	request, err := http.NewRequest("GET", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	return s.httpConnPool[address].Do(request)
}

/*
	GPRC
*/

// ping method starts a GRPC client and make ping to the destination address.
// Ping returns the input ping message as an output, although we don't care about it.
func (s *Server) ping(destination string, payload uint64) error {
	var err error
	c := NewPingServerClient(s.grpcConnPool[destination])

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.Ping(ctx, &PingMessage{})
	if err != nil {
		log.Infof("could not ping: %v", err)
		return err
	}

	prometheus.RecordRequestSent(destination, payload)
	return nil
}
