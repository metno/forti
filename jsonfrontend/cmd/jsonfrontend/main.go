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
	upstream := flag.String("upstream", "localhost:5051", "get data from the given grpc server")
	metricsPort := flag.Int("metricsPort", 9090, "serve metrics on the given port")
	profilePort := flag.Int("profilePort", 0, "serve cpu profiles on the given port")
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
