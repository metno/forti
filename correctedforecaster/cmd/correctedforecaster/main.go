package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/fileblob"

	"gitlab.met.no/forti/f2/correctedforecaster/internal/download"
	"gitlab.met.no/forti/f2/correctedforecaster/internal/server"
)

func main() {
	upstream := flag.String("upstream", "localhost:5052", "get data from the given grpc server")
	bucket := flag.String("download-from", "", "download data from the given bucket")
	workdir := flag.String("workdir", "/data/", "use files in the given directory")
	port := flag.Int("port", 5051, "Listen port for incoming grpc requests.")
	stats := flag.Bool("serve-stats", false, "serve prometheus stats")

	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	topographyFiles, err := download.Get(ctx, *bucket, *workdir)
	if err != nil {
		log.Fatalf("Unable to get topography files: %s", err)
	}

	if *stats {
		go serveStats()
	}

	log.Println("ready")
	log.Fatalln(server.Run(*upstream, *port, topographyFiles))
}

func serveStats() {
	log.Println("serving prometheus stats on http://localhost:8080/metrics")
	http.Handle("/metrics", promhttp.Handler())
	log.Println(http.ListenAndServe(":8080", nil))
}
