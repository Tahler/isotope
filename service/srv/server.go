package srv

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"runtime"

	"github.com/Tahler/isotope/convert/pkg/graph"
	"github.com/Tahler/isotope/convert/pkg/graph/script"
	"github.com/Tahler/isotope/convert/pkg/graph/svctype"
	log "github.com/fortio/fortio/log"
	grpc "google.golang.org/grpc"
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

	grpcServer   *grpc.Server
	grpcPort     string
	grpcConnPool map[string]*grpc.ClientConn

	graph *graph.ServiceGraph
}

func NewServer(name string) (*Server, error) {
	var err error

	s := &Server{
		name:         name,
		httpServer:   new(http.Server),
		grpcServer:   new(grpc.Server),
		grpcPort:     fmt.Sprintf("%d", ServiceGRPCPort),
		grpcConnPool: make(map[string]*grpc.ClientConn),
		graph:        &graph.ServiceGraph{},
	}

	s.graph, err = serviceGraphFromYAMLFile(*configFile)
	if err != nil {
		return nil, err
	}

	// Create grpc connection pool to avoid creating connections in every request.
	err = s.createGrpcConnectionPool()
	if err != nil {
		return nil, err
	}

	// HTTP server initialization.
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

func (s *Server) createGrpcConnectionPool() error {
	service, err := extractService(*s.graph, s.name)
	if err != nil {
		return err
	}

	if service.Type == svctype.ServiceGRPC {

		for _, task := range service.Script {
			switch cmd := task.(type) {
			case script.RequestCommand:
				addr := cmd.ServiceName + ":" + s.grpcPort
				conn, err := grpc.Dial(addr, grpc.WithInsecure())
				if err != nil {
					log.Fatalf("Could not create GRPC connection: %v", err)
				}
				s.grpcConnPool[cmd.ServiceName] = conn

			case script.ConcurrentCommand:
				for _, subcmd := range cmd {
					sc := subcmd.(script.RequestCommand)
					addr := sc.ServiceName + ":" + s.grpcPort
					conn, err := grpc.Dial(addr, grpc.WithInsecure())
					if err != nil {
						log.Fatalf("Could not create GRPC connection: %v", err)
					}
					s.grpcConnPool[sc.ServiceName] = conn
				}

			}
		}
	}

	return nil
}
