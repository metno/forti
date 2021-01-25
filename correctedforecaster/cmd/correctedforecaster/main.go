package main

import (
	"context"
	"flag"
	"log"
	"time"

	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/fileblob"

	"gitlab.met.no/forti/f2/correctedforecaster/internal/download"
	"gitlab.met.no/forti/f2/correctedforecaster/internal/server"
)

func main() {
	upstream := flag.String("upstream", "localhost:5052", "get data from the given grpc server")
	bucket := flag.String("download-from", "", "download data from the given bucket")
	workdir := flag.String("workdir", "/data/", "use files in the given directory")
	port := flag.Int("port", 5051, "Listen port for incoming grpc requests.")

	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	topographyFiles, err := download.Get(ctx, *bucket, *workdir)
	if err != nil {
		log.Fatalf("Unable to get topography files: %s", err)
	}

	log.Println("ready")
	log.Fatalln(server.Run(*upstream, *port, topographyFiles))
}
