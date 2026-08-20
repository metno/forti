package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/s3blob"

	"github.com/metno/forti/rawdataforecaster/internal/server"
	"github.com/metno/forti/rawdataforecaster/internal/server/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	confFile := flag.String("config", "config.json", "Path to JSON configuration file (defines data source, areas, loader strategy)")
	stats := flag.Bool("serve-stats", false, "Enable Prometheus metrics endpoint on :8080/metrics")
	port := flag.Int("port", 5052, "gRPC server listen port")

	flag.Parse()

	conf, err := getConfig(*confFile)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("use %s storage strategy", conf.Loader.Type)

	if *stats {
		go serveStats()
	}

	log.Fatalln(server.Run(conf, *port))
}

func getConfig(file string) (*config.Configuration, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("unable to open config file: %w", err)
	}
	defer f.Close()

	var out config.Configuration
	if err := json.NewDecoder(f).Decode(&out); err != nil {
		return nil, fmt.Errorf("unable to parse config file: %w", err)
	}

	return &out, nil
}

func serveStats() {
	log.Println("serving prometheus stats on http://localhost:8080/metrics")
	http.Handle("/metrics", promhttp.Handler())
	log.Println(http.ListenAndServe(":8080", nil))
}
