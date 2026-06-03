package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/metno/forti/moxfrontend/internal/server"
)

func main() {
	upstream := flag.String("upstream", "localhost:5052", "get data from the given grpc server")
	flag.Parse()

	server, err := server.New(*upstream)
	if err != nil {
		log.Fatalln(err)
	}

	http.Handle("/", server)

	log.Println("ready")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
