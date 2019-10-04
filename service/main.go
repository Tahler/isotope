package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Tahler/isotope/service/srv"
	"istio.io/fortio/log"
)

func main() {
	// Check flags and env. variables.
	flag.Parse()
	name, ok := os.LookupEnv(srv.ServiceNameEnvKey)
	if !ok {
		log.Fatalf(`env var "%s" is not set`, srv.ServiceNameEnvKey)
	}

	log.SetLogLevel(log.Info)

	// Start servers.
	server := srv.NewServer(name)
	err := server.Start()
	if err != nil {
		log.Fatalf("Can't start service: %v", err)
	}

	AwaitTermSignal(server.Stop)
	log.Infof("Stopping service, about to exit...")
}

func AwaitTermSignal(closeFn func() error) {

	signals := make(chan os.Signal, 1)
	// sigint: Ctrl-C, sigterm: kill command
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// block main and wait for a signal
	sig := <-signals
	fmt.Printf("Signal received: %v\n", sig)

	_ = closeFn()
}
