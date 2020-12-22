package main

import (
	"flag"
	"log"
	"net/http"
	"strings"

	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/fileblob"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.met.no/forti/f2/simpleforecaster/internal/server"
)

func main() {
	bucket := flag.String("blob", "file:///home/vegardb/local/forti/collected", "Read forecasts from the given bucket")
	groups := flag.String("groups", "nordic", "Serve the given groups")
	workdir := flag.String("workdir", "", "use the given folder as a workdir")
	stats := flag.Bool("serve-stats", false, "serve prometheus stats")
	flag.Parse()

	conf := server.Configuration{
		Bucket:  *bucket,
		Groups:  strings.Split(*groups, ","),
		Workdir: *workdir,
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
