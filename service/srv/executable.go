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
	"github.com/Tahler/isotope/convert/pkg/graph/svctype"
	"github.com/Tahler/isotope/service/srv/prometheus"
	multierror "github.com/hashicorp/go-multierror"
	"istio.io/fortio/log"
)

func execute(step interface{}, forwardableHeader http.Header, serviceTypes map[string]svctype.ServiceType) (err error) {
	switch cmd := step.(type) {
	case script.SleepCommand:
		executeSleepCommand(cmd)
	case script.RequestCommand:
		err = executeRequestCommand(cmd, forwardableHeader, serviceTypes)
	case script.ConcurrentCommand:
		err = executeConcurrentCommand(cmd, forwardableHeader, serviceTypes)
	default:
		log.Fatalf("unknown command type in script: %T", cmd)
	}
	return
}

func executeSleepCommand(cmd script.SleepCommand) {
	time.Sleep(time.Duration(cmd))
}

// Execute sends an HTTP request to another service. Assumes DNS is available
// which maps exe.ServiceName to the relevant URL to reach the service.
func executeRequestCommand(
	cmd script.RequestCommand,
	forwardableHeader http.Header,
	serviceTypes map[string]svctype.ServiceType) (err error) {
	destName := cmd.ServiceName
	// destType, ok := serviceTypes[destName]
	// if !ok {
	// 	err = fmt.Errorf("service %s does not exist", destName)
	// 	return
	// }
	response, err := sendRequest(destName, uint64(cmd.Size), forwardableHeader)
	if err != nil {
		return
	}
	prometheus.RecordRequestSent(destName, uint64(cmd.Size))
	if response.StatusCode == 200 {
		log.Debugf("%s responded with %s", destName, response.Status)
	} else {
		log.Errf("%s responded with %s", destName, response.Status)
	}
	if response.StatusCode == http.StatusInternalServerError {
		err = fmt.Errorf("service %s responded with %s", destName, response.Status)
	}

	// Necessary for reusing HTTP/1.x "keep-alive" TCP connections.
	// https://golang.org/pkg/net/http/#Response
	readAllAndClose(response.Body)

	return
}

func readAllAndClose(r io.ReadCloser) {
	_, _ = io.Copy(ioutil.Discard, r)
	r.Close()
}

// executeConcurrentCommand calls each command in exe.Commands asynchronously
// and waits for each to complete.
func executeConcurrentCommand(cmd script.ConcurrentCommand, forwardableHeader http.Header, serviceTypes map[string]svctype.ServiceType) (errs error) {
	numSubCmds := len(cmd)
	wg := sync.WaitGroup{}
	wg.Add(numSubCmds)
	for _, subCmd := range cmd {
		go func(step interface{}) {
			defer wg.Done()

			err := execute(step, forwardableHeader, serviceTypes)
			if err != nil {
				errs = multierror.Append(errs, err)
			}
		}(subCmd)
	}
	wg.Wait()
	return
}

// HTTP

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

// GPRC ?
