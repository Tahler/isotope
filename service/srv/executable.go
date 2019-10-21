package srv

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/Tahler/isotope/convert/pkg/graph/script"
	"github.com/Tahler/isotope/service/srv/prometheus"
	"github.com/fortio/fortio/log"
	multierror "github.com/hashicorp/go-multierror"
)

func (s *Server) execute(step interface{}) (err error) {
	switch cmd := step.(type) {
	case script.SleepCommand:
		time.Sleep(time.Duration(cmd))
	case script.RequestCommand:
		err = s.executeRequestCommand(cmd)
	case script.ConcurrentCommand:
		err = s.executeConcurrentCommand(cmd)
	default:
		log.Fatalf("unknown command type in script: %T", cmd)
	}
	return
}

// Execute sends an HTTP request to another service. Assumes DNS is available
// which maps exe.ServiceName to the relevant URL to reach the service.
func (s *Server) executeRequestCommand(cmd script.RequestCommand) (err error) {
	destination := cmd.ServiceName

	response, err := s.sendRequest(destination, uint64(cmd.Size))
	if err != nil {
		return
	}

	prometheus.RecordRequestSent(destination, uint64(cmd.Size))
	if response.StatusCode != 200 {
		// 	log.Debugf("%s responded with %s", destination, response.Status)
		// } else {
		log.Errf("%s responded with %s", destination, response.Status)
	}
	if response.StatusCode == http.StatusInternalServerError {
		err = fmt.Errorf("service %s responded with %s", destination, response.Status)
	}

	// Necessary for reusing HTTP/1.x "keep-alive" TCP connections.
	// https://golang.org/pkg/net/http/#Response
	_, _ = io.Copy(ioutil.Discard, response.Body)
	response.Body.Close()

	return
}

// executeConcurrentCommand calls each command in exe.Commands asynchronously
// and waits for each to complete.
func (s *Server) executeConcurrentCommand(cmd script.ConcurrentCommand) (errs error) {
	numSubCmds := len(cmd)
	wg := sync.WaitGroup{}
	wg.Add(numSubCmds)
	for _, subCmd := range cmd {
		go func(step interface{}) {
			defer wg.Done()
			err := s.executeRequestCommand(step.(script.RequestCommand))
			if err != nil {
				errs = multierror.Append(errs, err)
			}
		}(subCmd)
	}
	wg.Wait()
	return
}

func (s *Server) sendRequest(address string, payloadSize uint64) (*http.Response, error) {
	url := fmt.Sprintf("http://%s:%v", address, ServiceHTTPPort)

	// Build request
	payload := make([]byte, payloadSize)
	request, err := http.NewRequest("GET", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	return s.httpConnPool[address].Do(request)
}
