package main

import (
	"flag"
	"log"
	"net/http"

	"gitlab.met.no/forti/f2/xmlfrontend/internal/server"
	"gitlab.met.no/forti/f2/xmlfrontend/internal/server/config"
)

func main() {
	upstream := flag.String("upstream", "localhost:50051", "get data from the given grpc server")
	configFile := flag.String("config", "xmlformat.json", "read json formatting instructions from the given file")
	flag.Parse()

	if err := config.Initialize(*configFile); err != nil {
		log.Fatalf("unable to read configuration: %s", err)
	}

	server, err := server.New(*upstream)
	if err != nil {
		log.Fatalln(err)
	}

	http.Handle("/", server)

	log.Println("ready")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
