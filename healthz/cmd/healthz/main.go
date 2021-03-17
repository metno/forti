package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"

	"gitlab.met.no/forti/f2/healthz/internal/health"
	"gitlab.met.no/forti/f2/healthz/internal/health/json/config"
	"gitlab.met.no/forti/f2/healthz/internal/status"
)

func main() {
	runServer := flag.Bool("run-server", false, "Run a server on port 8080, continously serving status.")
	configFile := flag.String("config", "/etc/forti/checks.json", "Read check configuration from the given file.")
	upstreamGRPC := flag.String("upstream-grpc-host", "", "Check status against the given grpc server")
	flag.Parse()

	conf, err := config.Read(*configFile)
	if err != nil {
		log.Fatalln(err)
	}

	if *runServer {
		log.Fatal(serveHTTP(conf, *upstreamGRPC))
	} else {
		checkOnce(conf)
	}
}

func serveHTTP(conf *config.CheckConfiguration, upstreamGRPC string) error {
	checker := health.NewChecker(conf)
	http.HandleFunc("/healthz", checker.ServeSimple)
	http.HandleFunc("/healthz/full", checker.ServeJSON)

	if upstreamGRPC != "" {
		statusFetcher, err := status.NewFetcher(upstreamGRPC)
		if err != nil {
			return err
		}
		http.Handle("/status", statusFetcher)
	}

	address := ":8080"
	log.Println("ok. Serving data on " + address)
	return http.ListenAndServe(address, nil)
}

func checkOnce(conf *config.CheckConfiguration) {
	checker := health.NewChecker(conf)

	result := checker.Check()
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		panic(err)
	}
	if !result.OK {
		os.Exit(1)
	}
}
