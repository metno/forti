package main

import (
	"flag"
	"log"
	"net/http"
	"strings"

	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/fileblob"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server"
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/config"
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast/dataset/values/blob"
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast/dataset/values/file"
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast/dataset/values/memory"
)

func main() {
	bucket := flag.String("blob", "file:///home/vegardb/local/forti/collected", "Read forecasts from the given bucket")
	areas := flag.String("areas", "nordic", "Serve the given areas")
	storage := flag.String("storage", "memory", "Use the given data storage strategy. Available: memory, file and blob.")
	stats := flag.Bool("serve-stats", false, "serve prometheus stats")
	flag.Parse()

	conf := config.Configuration{
		Bucket: *bucket,
		Areas:  strings.Split(*areas, ","),
	}
	switch *storage {
	case "memory":
		conf.ValueDownloadFunction = memory.Download
	case "file":
		conf.ValueDownloadFunction = file.Download
	case "blob":
		conf.ValueDownloadFunction = blob.Download
	default:
		log.Fatalf("invalid strategy: %s", *storage)
	}
	log.Printf("use %s storage strategy", *storage)

	if *stats {
		go serveStats()
	}

	log.Fatalln(server.Run(&conf))
}

func serveStats() {
	http.Handle("/metrics", promhttp.Handler())
	log.Println(http.ListenAndServe(":8080", nil))
}
