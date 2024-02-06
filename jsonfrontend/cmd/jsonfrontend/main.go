package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"

	"gitlab.met.no/forti/f2/jsonfrontend/internal/server"
	"gitlab.met.no/forti/f2/jsonfrontend/internal/server/config"
	"gitlab.met.no/forti/f2/jsonfrontend/internal/server/metrics"
)

func main() {
	upstream := flag.String("upstream", "localhost:5051", "get data from the given grpc server")
	metricsPort := flag.Int("metricsPort", 9090, "serve metrics on the given port")
	configFile := flag.String("config", "jsonformat.json", "read json formatting instructions from the given file")
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

	go func() {
		r := http.NewServeMux()
		r.HandleFunc("/debug/pprof/profile", pprof.Profile)
		log.Fatal(http.ListenAndServe(":6080", r))
	}()

	http.Handle("/", server)
	log.Println("ready")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
