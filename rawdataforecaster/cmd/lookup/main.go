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
	address := flag.String("address", "localhost:5052", "Server to connect to")
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

	request := internalprotocol.GetForecastRequest{
		Latitude:  float32(*latitude),
		Longitude: float32(*longitude),
	}

	forecast, err := c.GetForecast(ctx, &request)
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	if forecast.ForecastStatus != internalprotocol.ForecastStatus_OK {
		log.Fatalln(forecast.ForecastStatus)
	}

	fmt.Printf("updated at: %v\n", forecast.ForecastMeta.UpdatedAt.AsTime())
	fmt.Printf("next update: %v\n", forecast.ForecastMeta.NextUpdate.AsTime())
	for _, meta := range forecast.ParameterMeta {
		fmt.Printf("%s (%s):\n", meta.Parameter, meta.Units)
		for i, t := range meta.Times {
			value := forecast.Data[i+int(meta.SliceFrom)]
			fmt.Printf("  %v: %0.1f\n", t.AsTime(), value)
		}
	}
}
