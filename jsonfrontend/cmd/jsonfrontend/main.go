package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.met.no/forti/f2/jsonfrontend/internal/server"
	"gitlab.met.no/forti/f2/jsonfrontend/internal/server/config"
)

func main() {
	upstream := flag.String("upstream", "localhost:5051", "get data from the given grpc server")
	configFile := flag.String("config", "jsonformat.json", "read json formatting instructions from the given file")
	flag.Parse()

	if err := config.Initialize(*configFile); err != nil {
		log.Fatalf("unable to read configuration: %s", err)
	}

	server, err := server.New(*upstream)
	if err != nil {
		log.Fatalln(err)
	}

	http.Handle("/", server)
	http.Handle("/metrics", promhttp.Handler())

	log.Println("ready")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
