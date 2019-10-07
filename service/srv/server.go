package srv

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"runtime"

	"github.com/Tahler/isotope/convert/pkg/graph"
	grpc "google.golang.org/grpc"
	log "istio.io/fortio/log"
)

var (
	maxIdleConnectionsPerHostFlag = flag.Int(
		"max-idle-connections-per-host", 0,
		"maximum number of connections to keep open per host")

	configFile = flag.String(
		"config-file", "/etc/config/service-graph.yaml",
		"the full path with file name which contains the configuration file")
)

type Server struct {
	name        string
	http_server *http.Server
	grpc_server *grpc.Server
	grpc_port   string

	graph *graph.ServiceGraph
}

func NewServer(name string) *Server {
	var err error

	s := &Server{
		name:        name,
		http_server: new(http.Server),
		grpc_server: new(grpc.Server),
		grpc_port:   fmt.Sprintf(":%d", ServiceGRPCPort),
		graph:       &graph.ServiceGraph{},
	}

	s.graph, err = serviceGraphFromYAMLFile(*configFile)
	if err != nil {
		return nil
	}

	defaultHandler, err := HandlerFromServiceGraphYAML(s.name, *s.graph)
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

func (s *Server) Start() error {

	// Start GRPC server in a different goroutine.
	lis, err := net.Listen("tcp", s.grpc_port)
	if err != nil {
		log.Infof("failed to listen GRPC: %v", err)
		return err
	}

	s.grpc_server = grpc.NewServer()
	RegisterPingServerServer(s.grpc_server, s)
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
