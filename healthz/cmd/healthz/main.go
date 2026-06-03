package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/metno/forti/healthz/internal/health"
	"github.com/metno/forti/healthz/internal/health/config"
	"github.com/metno/forti/healthz/internal/status"
)

func main() {
	runServer := flag.Bool("run-server", false, "Run a server on port 8080, continously serving status.")
	configFile := flag.String("config", "/etc/forti/probes.json", "Read probe configuration from the given file.")
	upstreamGRPC := flag.String("upstream-grpc-host", "", "Check status against the given grpc server")
	onlyCheckConfig := flag.Bool("only-check-config", false, "Do not run checks. Only verify configuration, then exit")

	flag.Parse()

	conf, err := config.Read(*configFile)
	if err != nil {
		log.Fatalln(err)
	}

	if errs := conf.Problems(); len(errs) != 0 {
		log.Println("Problems in configuration:")
		for _, e := range errs {
			log.Printf("* %s", e)
		}
		log.Fatalln("Cannot proceed")
	}
	if *onlyCheckConfig {
		log.Println("ok")
		return
	}

	if *runServer {
		log.Fatal(serveHTTP(conf, *upstreamGRPC))
	} else {
		checkOnce(conf)
	}
}

func serveHTTP(conf *config.ProbeConfiguration, upstreamGRPC string) error {
	h := health.New(conf)
	h.Start()

	http.HandleFunc("/healthz", h.ServeSimple)
	http.HandleFunc("/healthz/full", h.ServeJSON)

	http.HandleFunc("/healthz/{type}", h.ServeTypeSimple)
	http.HandleFunc("/healthz/{type}/full", h.ServeTypeJSON)

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

func checkOnce(conf *config.ProbeConfiguration) {
	h := health.New(conf)
	h.Probe()
	lastProbe, isHealthy := h.Health()

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(lastProbe); err != nil {
		panic(err)
	}
	if !isHealthy {
		os.Exit(1)
	}
}
