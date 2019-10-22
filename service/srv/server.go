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
	log "github.com/fortio/fortio/log"
	grpc "google.golang.org/grpc"
)

var (
	configFile = flag.String(
		"config-file", "/etc/config/service-graph.yaml",
		"the full path with file name which contains the configuration file")
)

type Server struct {
	name         string
	httpServer   *http.Server
	httpConnPool map[string]*http.Client

	grpcServer   *grpc.Server
	grpcPort     string
	grpcConnPool map[string]*grpc.ClientConn

	graph *graph.ServiceGraph
	tasks []*Task
}

func NewServer(name string) (*Server, error) {
	var err error

	s := &Server{
		name:         name,
		httpServer:   new(http.Server),
		httpConnPool: make(map[string]*http.Client),
		grpcServer:   new(grpc.Server),
		grpcPort:     fmt.Sprintf("%d", ServiceGRPCPort),
		grpcConnPool: make(map[string]*grpc.ClientConn),
		graph:        &graph.ServiceGraph{},
	}

	s.graph, err = serviceGraphFromYAMLFile(*configFile)
	if err != nil {
		return nil, err
	}

	// Create grpc and http connection pools to avoid creating connections in every request.
	err = s.createTasksAndConnectionPools()
	if err != nil {
		return nil, err
	}

	// HTTP server initialization.
	defaultHandler, err := HandlerFromServiceGraphYAML(s.name, *s.graph)
	if err != nil {
		log.Fatalf("%s", err)
	}

	mux := s.newApiHttp(defaultHandler)
	s.httpServer.Addr = fmt.Sprintf(":%d", ServiceHTTPPort)
	s.httpServer.Handler = mux

	setMaxProcs()
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

func setMaxProcs() {
	numCPU := runtime.NumCPU()
	maxProcs := runtime.GOMAXPROCS(0)
	if maxProcs < numCPU {
		log.Infof("setting GOMAXPROCS to %v (previously %v)", numCPU, maxProcs)
		runtime.GOMAXPROCS(numCPU)
	}
}

func (s *Server) createTasksAndConnectionPools() error {
	service, err := extractService(*s.graph, s.name)
	if err != nil {
		return err
	}

	for _, ta := range service.Script {
		var t *Task
		switch cmd := ta.(type) {
		case script.RequestCommand:

			// Create connection pools.
			addr := cmd.ServiceName + ":" + s.grpcPort
			conn, err := grpc.Dial(addr, grpc.WithInsecure())
			if err != nil {
				log.Fatalf("Could not create GRPC connection: %v", err)
			}
			s.grpcConnPool[cmd.ServiceName] = conn
			s.httpConnPool[cmd.ServiceName] = &http.Client{}

			// Create tasks list. Only 1 task here.
			url := fmt.Sprintf("http://%s:%v", cmd.ServiceName, ServiceHTTPPort)
			t = newTask(cmd, service.Type, cmd.ServiceName, url, uint64(cmd.Size))
			s.tasks = append(s.tasks, t)

		case script.ConcurrentCommand:
			for _, subcmd := range cmd {
				sc := subcmd.(script.RequestCommand)

				// Create connection pools.
				addr := sc.ServiceName + ":" + s.grpcPort
				conn, err := grpc.Dial(addr, grpc.WithInsecure())
				if err != nil {
					log.Fatalf("Could not create GRPC connection: %v", err)
				}
				s.grpcConnPool[sc.ServiceName] = conn

				// Create tasks list. +1 task.
				url := fmt.Sprintf("http://%s:%v", sc.ServiceName, ServiceHTTPPort)
				t = newTask(sc, service.Type, sc.ServiceName, url, uint64(sc.Size))
				s.tasks = append(s.tasks, t)
			}

		case script.SleepCommand:
			t = newTask(cmd, service.Type, "", "", uint64(0))
			s.tasks = append(s.tasks, t)
		}
	}

	return nil
}
