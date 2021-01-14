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
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast/fortidb"
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast/fortidb/values/file"
)

func main() {
	bucket := flag.String("blob", "file:///home/vegardb/local/forti/collected", "Read forecasts from the given bucket")
	areas := flag.String("areas", "nordic", "Serve the given areas")
	useFiles := flag.Bool("use-files", false, "Store data in local file system instead of in memory. This is meant for testing.")
	stats := flag.Bool("serve-stats", false, "serve prometheus stats")
	flag.Parse()

	if *useFiles {
		log.Println("storing data in file system instead of memory")
		fortidb.SetDownloadFunction(file.Download)
	}

	conf := server.Configuration{
		Bucket: *bucket,
		Areas:  strings.Split(*areas, ","),
	}

	if *stats {
		go serveStats()
	}

	log.Fatalln(server.Run(&conf))
}

func serveStats() {
	http.Handle("/metrics", promhttp.Handler())
	log.Println(http.ListenAndServe(":8080", nil))
}
