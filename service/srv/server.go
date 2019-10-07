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
	name       string
	httpServer *http.Server
	grpcServer *grpc.Server
	grpcPort   string

	graph *graph.ServiceGraph
}

func NewServer(name string) (*Server, error) {
	var err error

	s := &Server{
		name:       name,
		httpServer: new(http.Server),
		grpcServer: new(grpc.Server),
		grpcPort:   fmt.Sprintf("%d", ServiceGRPCPort),
		graph:      &graph.ServiceGraph{},
	}

	s.graph, err = serviceGraphFromYAMLFile(*configFile)
	if err != nil {
		return nil, err
	}

	defaultHandler, err := HandlerFromServiceGraphYAML(s.name, *s.graph)
	if err != nil {
		log.Fatalf("%s", err)
	}

	mux := newApiHttp(defaultHandler)
	s.httpServer.Addr = fmt.Sprintf(":%d", ServiceHTTPPort)
	s.httpServer.Handler = mux

	setMaxProcs()
	setMaxIdleConnectionsPerHost(*maxIdleConnectionsPerHostFlag)
	return s, nil
}

func (s *Server) Start() error {

	// Start GRPC server in a different goroutine.
	grcpAddress := ":" + s.grpcPort
	lis, err := net.Listen("tcp", grcpAddress)
	if err != nil {
		log.Infof("failed to listen GRPC: %v", err)
		return err
	}

	s.grpcServer = grpc.NewServer()
	RegisterPingServerServer(s.grpcServer, s)
	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve GRPC: %v", err)
		}
		log.Infof("listening GRPC on port %v\n", s.grpcPort)
	}()

	// Start HTTP server on the main thread.
	log.Infof("listening HTTP on port %v\n", s.httpServer.Addr)
	err = s.httpServer.ListenAndServe()
	if err != nil {
		log.Fatalf("%s", err)
	}

	return nil
}

func (s *Server) Stop() error {
	// Stopping HTTP server
	log.Infof("Stopping HTTP server...")
	if err := s.httpServer.Shutdown(context.Background()); err != nil {
		log.Infof("Unable to stop HTTP server: %v", err)
		return err
	}

	// Stopping GRPC server
	log.Infof("Stopping GRPC server...")
	s.grpcServer.GracefulStop()

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
