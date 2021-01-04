package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"gitlab.met.no/forti/f2/internalprotocol"
	"google.golang.org/grpc"
)

func main() {
	latitude := flag.Float64("lat", 59, "latitude to query")
	longitude := flag.Float64("lon", 11, "longitude to query")
	address := flag.String("address", "localhost:50051", "Server to connect to")
	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.Dial(*address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := internalprotocol.NewForecasterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	request := internalprotocol.Location{
		Latitude:  float32(*latitude),
		Longitude: float32(*longitude),
	}

	r, err := c.GetForecast(ctx, &request)
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	// b, err := json.MarshalIndent(r, "", "  ")
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// fmt.Println(string(b))

	fmt.Printf("updated at: %v\n", r.Meta.UpdatedAt.AsTime())
	fmt.Printf("next update: %v\n", r.Meta.NextUpdate.AsTime())
	for _, data := range r.Data {
		for _, meta := range data.ParameterMeta {
			fmt.Printf("%s (%s):\n", meta.Parameter, meta.Units)
			for i, t := range meta.Times {
				value := data.Data[i+int(meta.SliceFrom)]
				fmt.Printf("  %v: %0.1f\n", t.AsTime(), value)
			}
		}
	}
}
