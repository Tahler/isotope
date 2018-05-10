package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Tahler/service-grapher/pkg/graph"

	multierror "github.com/hashicorp/go-multierror"
)

func execute(step interface{}, forwardableHeader http.Header) (err error) {
	switch cmd := step.(type) {
	case graph.SleepCommand:
		executeSleepCommand(cmd)
	case graph.RequestCommand:
		err = executeRequestCommand(cmd, forwardableHeader)
	case graph.ConcurrentCommand:
		err = executeConcurrentCommand(cmd, forwardableHeader)
	default:
		log.Fatalf("unknown command type in script: %T", cmd)
	}
	return
}

func executeSleepCommand(cmd graph.SleepCommand) {
	time.Sleep(cmd.Duration)
}

// Execute sends an HTTP request to another service. Assumes DNS is available
// which maps exe.ServiceName to the relevant URL to reach the service.
func executeRequestCommand(
	cmd graph.RequestCommand, forwardableHeader http.Header) (err error) {
	url := fmt.Sprintf("http://%s:%v", cmd.ServiceName, port)
	request, err := buildRequest(
		cmd.HTTPMethod, url, cmd.RequestSettings.Size, forwardableHeader)
	if err != nil {
		return
	}
	log.Printf(
		"Sending %s request to %s (%s)", cmd.HTTPMethod, cmd.ServiceName, url)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return
	}
	log.Printf("%s responded with %s", cmd.ServiceName, response.Status)
	if response.StatusCode == http.StatusInternalServerError {
		err = fmt.Errorf(
			"service %s responded with %s", cmd.ServiceName, response.Status)
	}
	return
}

func buildRequest(
	method graph.HTTPMethod, url string, size int64, requestHeader http.Header) (
	request *http.Request, err error) {
	payload := make([]byte, size, size)
	request, err = http.NewRequest(string(method), url, bytes.NewBuffer(payload))
	if err != nil {
		return
	}
	copyHeader(request, requestHeader)
	return
}

func copyHeader(request *http.Request, header http.Header) {
	for key, values := range header {
		request.Header[key] = values
	}
}

// executeConcurrentCommand calls each command in exe.Commands asynchronously
// and waits for each to complete.
func executeConcurrentCommand(
	cmd graph.ConcurrentCommand, forwardableHeader http.Header) error {
	wg := sync.WaitGroup{}
	wg.Add(len(cmd.Commands))
	var errs *multierror.Error
	for _, subCmd := range cmd.Commands {
		go func(step interface{}) {
			err := execute(step, forwardableHeader)
			errs = multierror.Append(errs, err)
			wg.Done()
		}(subCmd)
	}
	wg.Wait()
	return errs
}
