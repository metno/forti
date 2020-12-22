package main

import (
	"flag"
	"log"
	"time"

	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/fileblob"
	"golang.org/x/net/context"

	"gitlab.met.no/forti/f2/correctedforecaster/internal/download"
	"gitlab.met.no/forti/f2/correctedforecaster/internal/server"
)

func main() {
	upstream := flag.String("upstream", "localhost:50051", "get data from the given grpc server")
	bucket := flag.String("download-from", "", "download data from the given bucket")
	workdir := flag.String("workdir", "/data/", "use files in the given directory")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	topographyFiles, err := download.Get(ctx, *bucket, *workdir)
	if err != nil {
		log.Fatalf("Unable to get topography files: %s", err)
	}

	log.Println("ready")
	log.Fatalln(server.Run(*upstream, topographyFiles))
}
