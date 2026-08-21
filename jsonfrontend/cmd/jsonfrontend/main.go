package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"

	"github.com/metno/forti/jsonfrontend/internal/server"
	"github.com/metno/forti/jsonfrontend/internal/server/config"
	"github.com/metno/forti/jsonfrontend/internal/server/metrics"
)

func main() {
	upstream := flag.String("upstream", "localhost:5051", "gRPC upstream server address (rawdataforecaster or correctedforecaster)")
	metricsPort := flag.Int("metricsPort", 9090, "Prometheus metrics HTTP port (0 to disable)")
	profilePort := flag.Int("profilePort", 0, "pprof CPU profiling HTTP port (0 to disable)")
	configFile := flag.String("config", "jsonformat.json", "Path to JSON configuration file (defines parameter mapping, time periods, HTTP headers)")
	flag.Parse()

	if err := config.Initialize(*configFile); err != nil {
		log.Fatalf("unable to read configuration: %s", err)
	}

	server, err := server.New(*upstream)
	if err != nil {
		log.Fatalln(err)
	}

	if *metricsPort != 0 {
		log.Printf("serving stats at port %d", *metricsPort)
		addr := fmt.Sprintf(":%d", *metricsPort)
		go func() {
			log.Fatalln(metrics.Serve(addr))
		}()
	}

	if *profilePort != 0 {
		log.Printf("serving cpu profiles at port %d", *profilePort)
		addr := fmt.Sprintf(":%d", *profilePort)
		go func() {
			r := http.NewServeMux()
			r.HandleFunc("/debug/pprof/profile", pprof.Profile)
			log.Fatalln(http.ListenAndServe(addr, r))
		}()
	}

	http.Handle("/", server)
	log.Println("ready")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
