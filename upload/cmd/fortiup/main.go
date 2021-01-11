package main

import (
	"context"
	"flag"
	"log"
	"time"

	"gitlab.met.no/forti/f2/upload/internal/nc/store"
	"gitlab.met.no/forti/f2/upload/internal/upload"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
)

func main() {
	group := flag.String("group", "", "group to store")
	version := flag.Int("version", 0, "version to store")
	bucket := flag.String("bucket", "file:///tmp/forti/", "store into the given bucket")
	ttl := flag.String("time-until-next", "1h", "expected time until next update is available")
	flag.Parse()
	files := flag.Args()

	if *group == "" {
		log.Fatalln("missing -group argument")
	}
	if *version == 0 {
		log.Fatalln("missing -version argument")
	}
	if len(files) == 0 {
		log.Fatalln("missing input files")
	}
	timeUntilNext, err := time.ParseDuration(*ttl)
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()

	b, err := blob.OpenBucket(ctx, *bucket)
	if err != nil {
		log.Fatalln(err)
	}
	u := upload.New(b)

	if err := store.Store(ctx, u, *group, *version, files, timeUntilNext); err != nil {
		log.Fatalln(err)
	}
}
