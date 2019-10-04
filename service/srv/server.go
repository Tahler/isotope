package srv

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"

	grpc "google.golang.org/grpc"
	log "istio.io/fortio/log"
)

const (
	promEndpoint    = "/metrics"
	defaultEndpoint = "/"
)

var (
	maxIdleConnectionsPerHostFlag = flag.Int(
		"max-idle-connections-per-host", 0,
		"maximum number of connections to keep open per host")

	configFile = flag.String(
		"config-file", "config/service-graph.yaml",
		"the full path with file name which contains the configuration file")
)

type Server struct {
	http_server *http.Server
	grpc_server *grpc.Server
	grpc_port   string
}

func NewServer() *Server {
	s := &Server{
		http_server: new(http.Server),
		grpc_server: new(grpc.Server),
		grpc_port:   fmt.Sprintf(":%d", ServiceGRPCPort),
	}

	serviceName, ok := os.LookupEnv(ServiceNameEnvKey)
	if !ok {
		log.Fatalf(`env var "%s" is not set`, ServiceNameEnvKey)
	}

	defaultHandler, err := HandlerFromServiceGraphYAML(*configFile, serviceName)
	if err != nil {
		log.Fatalf("%s", err)
	}

	mux := newApiHttp(defaultHandler)
	s.http_server.Addr = fmt.Sprintf(":%d", ServiceHTTPPort)
	s.http_server.Handler = mux

	setMaxProcs()
	setMaxIdleConnectionsPerHost(*maxIdleConnectionsPerHostFlag)
	return s
}

// TO REVIEW
func setMaxProcs() {
	numCPU := runtime.NumCPU()
	maxProcs := runtime.GOMAXPROCS(0)
	if maxProcs < numCPU {
		log.Infof("setting GOMAXPROCS to %v (previously %v)", numCPU, maxProcs)
		runtime.GOMAXPROCS(numCPU)
	}
}

// TO REVIEW
func setMaxIdleConnectionsPerHost(n int) {
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = n
}

func (s *Server) Start() error {

	// Start GRPC server in a different goroutine.
	lis, err := net.Listen("tcp", s.grpc_port)
	if err != nil {
		log.Infof("failed to listen GRPC: %v", err)
		return err
	}

	s.grpc_server = grpc.NewServer()
	RegisterPingServerServer(s.grpc_server, &Server{})
	go func() {
		if err := s.grpc_server.Serve(lis); err != nil {
			log.Fatalf("failed to serve GRPC: %v", err)
		}
		log.Infof("listening GRPC on port %v\n", s.grpc_port)
	}()

	// Start HTTP server on the main thread.
	log.Infof("listening HTTP on port %v\n", s.http_server.Addr)
	err = s.http_server.ListenAndServe()
	if err != nil {
		log.Fatalf("%s", err)
	}

	return nil
}

func (s *Server) Stop() error {
	// Stopping HTTP server
	log.Infof("Stopping HTTP server...")
	if err := s.http_server.Shutdown(context.Background()); err != nil {
		log.Infof("Unable to stop HTTP server: %v", err)
		return err
	}

	// Stopping GRPC server
	log.Infof("Stopping GRPC server...")
	s.grpc_server.GracefulStop()

	log.Infof("Done. Exiting...")
	return nil
}
