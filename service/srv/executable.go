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
	multierror "github.com/hashicorp/go-multierror"
	"istio.io/fortio/log"
)

func execute(step interface{}, forwardableHeader http.Header) (err error) {
	switch cmd := step.(type) {
	case script.SleepCommand:
		time.Sleep(time.Duration(cmd))
	case script.RequestCommand:
		err = executeRequestCommand(cmd, forwardableHeader)
	case script.ConcurrentCommand:
		err = executeConcurrentCommand(cmd, forwardableHeader)
	default:
		log.Fatalf("unknown command type in script: %T", cmd)
	}
	return
}

// Execute sends an HTTP request to another service. Assumes DNS is available
// which maps exe.ServiceName to the relevant URL to reach the service.
func executeRequestCommand(cmd script.RequestCommand, forwardableHeader http.Header) (err error) {
	destination := cmd.ServiceName

	response, err := sendRequest(destination, uint64(cmd.Size), forwardableHeader)
	if err != nil {
		return
	}

	prometheus.RecordRequestSent(destination, uint64(cmd.Size))
	if response.StatusCode == 200 {
		log.Debugf("%s responded with %s", destination, response.Status)
	} else {
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
func executeConcurrentCommand(cmd script.ConcurrentCommand, forwardableHeader http.Header) (errs error) {
	numSubCmds := len(cmd)
	wg := sync.WaitGroup{}
	wg.Add(numSubCmds)
	for _, subCmd := range cmd {
		go func(step interface{}) {
			defer wg.Done()
			err := executeRequestCommand(step.(script.RequestCommand), forwardableHeader)
			if err != nil {
				errs = multierror.Append(errs, err)
			}
		}(subCmd)
	}
	wg.Wait()
	return
}

func sendRequest(address string, payloadSize uint64, requestHeader http.Header) (*http.Response, error) {
	url := fmt.Sprintf("http://%s:%v", address, ServiceHTTPPort)

	// Build request
	payload := make([]byte, payloadSize)
	request, err := http.NewRequest("GET", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	// Copy header
	for key, values := range requestHeader {
		request.Header[key] = values
	}

	log.Debugf("sending request to %s ", url)
	return http.DefaultClient.Do(request)
}
