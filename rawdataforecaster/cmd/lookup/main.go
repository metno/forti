package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/metno/forti/internal/internalprotocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	latitude := flag.Float64("lat", 59, "latitude to query")
	longitude := flag.Float64("lon", 11, "longitude to query")
	altitude := flag.Float64("altitude", -1, "altitude to query; set to 0 for sea level, default -1 means not set")
	address := flag.String("address", "localhost:5052", "Server to connect to")
	parameter := flag.String("parameter", "", "Only show data for the given parameter")
	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.Dial(*address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
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
	if *altitude != -1 {
		request.Altitude = &internalprotocol.Altitude{
			Override: true,
			Value:    float32(*altitude),
		}
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
		if *parameter != "" && meta.Parameter != *parameter {
			continue
		}
		fmt.Printf("%s (%s):\n", meta.Parameter, meta.Units)
		for i, t := range meta.Times {
			value := forecast.Data[i+int(meta.SliceFrom)]
			fmt.Printf("  %v: %0.1f\n", t.AsTime(), value)
		}
	}
}
