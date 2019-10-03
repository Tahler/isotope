package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"runtime"

	"github.com/Tahler/isotope/service/srv"
	"github.com/Tahler/isotope/service/srv/prometheus"
	"google.golang.org/grpc"
	"istio.io/fortio/log"
)

const (
	promEndpoint    = "/metrics"
	defaultEndpoint = "/"
)

var (
	serviceGraphYAMLFilePath = path.Join(
		srv.ConfigPath, srv.ServiceGraphYAMLFileName)

	maxIdleConnectionsPerHostFlag = flag.Int(
		"max-idle-connections-per-host", 0,
		"maximum number of connections to keep open per host")
)

func main() {
	flag.Parse()

	log.SetLogLevel(log.Info)

	setMaxProcs()
	setMaxIdleConnectionsPerHost(*maxIdleConnectionsPerHostFlag)

	// Start GRPC server
	address := fmt.Sprintf(":%d", srv.ServiceGRPCPort)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	srv.RegisterPingServerServer(s, &srv.Server{})
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Start HTTP server
	serviceName, ok := os.LookupEnv(srv.ServiceNameEnvKey)
	if !ok {
		log.Fatalf(`env var "%s" is not set`, srv.ServiceNameEnvKey)
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

	addr := fmt.Sprintf(":%d", srv.ServiceHTTPPort)
	log.Infof("listening on port %v\n", srv.ServiceHTTPPort)
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
