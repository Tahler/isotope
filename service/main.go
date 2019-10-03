package main

import (
	"flag"

	"github.com/Tahler/isotope/service/srv"
	"github.com/bbva/qed/util"
	"istio.io/fortio/log"
)

func main() {
	flag.Parse()

	log.SetLogLevel(log.Info)

	server := srv.NewServer()
	err := server.Start()
	if err != nil {
		log.Fatalf("Can't start service: %v", err)
	}

	util.AwaitTermSignal(server.Stop)
	log.Infof("Stopping service, about to exit...")
}
