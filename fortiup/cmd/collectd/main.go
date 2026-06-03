package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/metno/forti/fortiup/internal/blob2blob/collector"
	"github.com/metno/forti/fortiup/internal/blob2blob/modelprovider"
)

func main() {
	in := flag.String("in", "file:///tmp/local/forti/", "Read forecasts from the given bucket")
	out := flag.String("out", "file:///tmp/local/forti/collected", "Write forecasts to the given bucket")
	flag.Parse()

	source, err := modelprovider.NewBlobClient(*in)
	if err != nil {
		log.Fatalln(err)
	}

	loaded := make(map[string]int)

	ctx := context.TODO()
	for {
		latest, err := source.Latest(ctx)
		if err != nil {
			log.Fatalln(err)
		}

		for _, dataset := range latest {
			if loaded[dataset.Group] < dataset.Version {
				log.Printf("upload %s/%d", dataset.Group, dataset.Version)
				if err := collector.Get(ctx, *in, *out, dataset.Group, dataset.Version); err != nil {
					log.Fatalln(err)
				}
				loaded[dataset.Group] = dataset.Version
				log.Println("done")
			}
		}

		time.Sleep(time.Minute)
	}
}
