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
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	grpc "google.golang.org/grpc"
)

var (
	maxIdleConnectionsPerHostFlag = flag.Int(
		"max-idle-connections-per-host", 0,
		"maximum number of connections to keep open per host")

	configFile = flag.String(
		"config-file",
		"/etc/config/service-graph.yaml",
		"the full path with file name which contains the configuration file")
)

type Server struct {
	name         string
	httpServer   *http.Server
	http2Server  *http.Server
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
		http2Server:  new(http.Server),
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

	setMaxProcs()
	mux := s.newApiHttp(defaultHandler)
	s.httpServer.Handler = mux
	s.httpServer.Addr = fmt.Sprintf(":%d", ServiceHTTPPort)

	h2s := &http2.Server{}
	mux2 := h2c.NewHandler(mux, h2s)
	s.http2Server.Handler = mux2
	s.http2Server.Addr = fmt.Sprintf(":%d", ServiceHTTP2Port)

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

	go func() {
		// Start HTTP2 server in a different goroutine
		log.Infof("listening HTTP on port %v\n", s.http2Server.Addr)
		err := s.http2Server.ListenAndServe()
		if err != nil {
			log.Fatalf("%s", err)
		}
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

	// Customize the Transport to have larger connection pool
	defaultRoundTripper := http.DefaultTransport
	defaultTransportPointer, ok := defaultRoundTripper.(*http.Transport)
	if !ok {
		panic(fmt.Sprintf("defaultRoundTripper not an *http.Transport"))
	}
	defaultTransport := *defaultTransportPointer // dereference it to get a copy of the struct that the pointer points to
	defaultTransport.MaxIdleConns = 100
	defaultTransport.MaxIdleConnsPerHost = 100

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
			s.httpConnPool[cmd.ServiceName] = &http.Client{
				Transport: &defaultTransport,
			}

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
				s.httpConnPool[sc.ServiceName] = &http.Client{
					Transport: &defaultTransport,
				}

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

	// Print tasks
	fmt.Printf("Service %s tasks:\n", s.name)
	for _, t := range s.tasks {
		fmt.Printf("%+v\n", t)
	}

	return nil
}
