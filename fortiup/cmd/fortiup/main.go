package main

import (
	"context"
	"flag"
	"log"
	"runtime"
	"time"

	"gitlab.met.no/forti/f2/fortiup/internal/nc/store"
	"gitlab.met.no/forti/f2/fortiup/internal/upload"
	"gitlab.met.no/forti/f2/fortiup/pkg/fortiblob"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/fileblob"
)

func main() {
	area := flag.String("area", "", "area to store")
	version := flag.Int("version", 0, "version to store")
	bucket := flag.String("bucket", "file:///tmp/forti/", "store into the given bucket")
	ttl := flag.String("time-until-next", "1h", "expected time until next update is available")
	wkt := flag.String("wkt", "", "specify geographic area for forecast")
	srs := flag.String("srs", "", "Spatial reference system for wkt")
	flag.Parse()
	files := flag.Args()

	if *area == "" {
		log.Fatalln("missing -area argument")
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

	meta := fortiblob.DatasetMeta{
		Area:          *area,
		Version:       *version,
		TimeUntilNext: timeUntilNext,
	}
	if *wkt != "" {
		if *srs == "" {
			log.Fatalln("-srs must be set if -wkt is set")
		}
		meta.GeographicExtent = &fortiblob.GeographicArea{
			WKT: *wkt,
			SRS: *srs,
		}
	}

	// TODO:
	// This application is currently targeted at ubuntu 24.04, which has a bug in its netcdf library (1.9.2) - see https://code.mpimet.mpg.de/boards/1/topics/14326
	// The bug should cause an avalanche of scary-looking log messages, although with seemingly no ill effect.
	// It seems the bug is fixed in libnetcdf 1.9.3.
	// Once this is available on the targetted platform, try to remove this locking, and reintroduct multi-threading in the store.Store function.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := store.Store(ctx, u, &meta, files); err != nil {
		log.Fatalln(err)
	}
}
