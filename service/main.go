package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"
	"runtime"

	"github.com/Tahler/isotope/convert/pkg/consts"
	"github.com/Tahler/isotope/service/pkg/srv"
	"github.com/Tahler/isotope/service/pkg/srv/prometheus"
	"istio.io/fortio/log"
)

const (
	promEndpoint    = "/metrics"
	defaultEndpoint = "/"
)

var (
	serviceGraphYAMLFilePath = path.Join(
		consts.ConfigPath, consts.ServiceGraphYAMLFileName)

	maxIdleConnectionsPerHostFlag = flag.Int(
		"max-idle-connections-per-host", 0,
		"maximum number of connections to keep open per host")
)

func main() {
	flag.Parse()

	log.SetLogLevel(log.Info)

	setMaxProcs()
	setMaxIdleConnectionsPerHost(*maxIdleConnectionsPerHostFlag)

	serviceName, ok := os.LookupEnv(consts.ServiceNameEnvKey)
	if !ok {
		log.Fatalf(`env var "%s" is not set`, consts.ServiceNameEnvKey)
	}

	defaultHandler, err := srv.HandlerFromServiceGraphYAML(
		serviceGraphYAMLFilePath, serviceName)
	if err != nil {
		log.Fatalf("%s", err)
	}

	err = serveWithPrometheus(defaultHandler)
	if err != nil {
		log.Fatalf("%s", err)
	}
}

func serveWithPrometheus(defaultHandler http.Handler) (err error) {
	log.Infof(`exposing Prometheus endpoint "%s"`, promEndpoint)
	http.Handle(promEndpoint, prometheus.Handler())

	log.Infof(`exposing default endpoint "%s"`, defaultEndpoint)
	http.Handle(defaultEndpoint, defaultHandler)

	addr := fmt.Sprintf(":%d", consts.ServicePort)
	log.Infof("listening on port %v\n", consts.ServicePort)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		return
	}
	return
}

func setMaxProcs() {
	numCPU := runtime.NumCPU()
	maxProcs := runtime.GOMAXPROCS(0)
	if maxProcs < numCPU {
		log.Infof("setting GOMAXPROCS to %v (previously %v)", numCPU, maxProcs)
		runtime.GOMAXPROCS(numCPU)
	}
}

func setMaxIdleConnectionsPerHost(n int) {
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = n
}
