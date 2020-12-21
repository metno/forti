package main

import (
	"context"
	"flag"
	"log"
	"time"

	"gitlab.met.no/forti/f2/upload/internal/collector"
)

func main() {
	group := flag.String("group", "", "group to collect")
	version := flag.Int("version", 0, "version to collect")
	in := flag.String("in", "file:///home/vegardb/local/forti/", "Read forecasts from the given bucket")
	out := flag.String("out", "file:///home/vegardb/local/forti/collected", "Write forecasts to the given bucket")
	timeout := flag.Int("timeout", 0, "Fail if not successful after the given time, in seconds")
	flag.Parse()

	ctx := context.Background()
	if *timeout != 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, time.Duration(*timeout)*time.Second)
		defer cancel()
	}

	if err := collector.Get(ctx, *in, *out, *group, *version); err != nil {
		log.Fatalln(err)
	}

}
