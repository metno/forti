package main

import (
	"context"
	"flag"
	"log"
	"time"

	"gitlab.met.no/forti/f2/fortiup/internal/blob2blob/collector"
)

func main() {
	area := flag.String("area", "", "group to collect")
	version := flag.Int("version", 0, "version to collect")
	in := flag.String("in", "file:///tmp/local/forti/", "Read forecasts from the given bucket")
	out := flag.String("out", "file:///tmp/local/forti/collected", "Write forecasts to the given bucket")
	timeout := flag.Int("timeout", 0, "Fail if not successful after the given time, in seconds")
	flag.Parse()

	ctx := context.Background()
	if *timeout != 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, time.Duration(*timeout)*time.Second)
		defer cancel()
	}

	if err := collector.Get(ctx, *in, *out, *area, *version); err != nil {
		log.Fatalln(err)
	}

}
