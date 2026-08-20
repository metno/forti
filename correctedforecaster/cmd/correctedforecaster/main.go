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

	"github.com/metno/forti/correctedforecaster/internal/download"
	"github.com/metno/forti/correctedforecaster/internal/server"
)

func main() {
	upstream := flag.String("upstream", "localhost:5052", "gRPC upstream server address (typically rawdataforecaster)")
	bucket := flag.String("download-from", "", "Bucket URL to download topography files from (e.g., 'file:///data/topography', 'azblob://...', empty to use -workdir only)")
	downloadTimeout := flag.Int("download-timeout", 240, "Timeout in seconds for downloading topography files from bucket")
	workdir := flag.String("workdir", "/data/", "Directory containing topography data files (used if -download-from is empty, or as download destination)")
	port := flag.Int("port", 5051, "gRPC server listen port")
	stats := flag.Bool("serve-stats", false, "Enable Prometheus metrics endpoint on :8080/metrics")

	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*downloadTimeout)*time.Second)
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
