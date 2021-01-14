package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"time"

	"gitlab.met.no/forti/f2/internalprotocol"
	"google.golang.org/grpc"
)

func main() {
	address := flag.String("address", "localhost:50051", "Server to connect to")
	threads := flag.Int("threads", 10, "number of threads")
	flag.Parse()

	timesCh := make(chan time.Duration)
	errCh := make(chan error)

	var duration time.Duration
	var count int
	go func() {
		for t := range timesCh {
			duration += t
			count++
		}
	}()
	var errors int
	go func() {
		for range errCh {
			errors++
		}
	}()

	stop := false
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	go func() {
		for range signals {
			stop = true
		}
	}()

	conn, err := grpc.Dial(*address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := internalprotocol.NewForecasterClient(conn)

	start := time.Now()
	var wait sync.WaitGroup
	for i := 0; i < *threads; i++ {
		wait.Add(1)
		go func() {
			for !stop {
				duration, err := randomRequest(c)
				if err != nil {
					errCh <- err
				}
				timesCh <- duration
			}
			wait.Done()
		}()
	}

	fmt.Println("running. Press CTRL-C to exit")

	wait.Wait()
	end := time.Now()

	timeToRun := end.Sub(start)

	fmt.Println()
	fmt.Printf("threads: %d\n", *threads)
	fmt.Printf("count: %d\n", count)
	fmt.Printf("errors: %v\n", errors)
	fmt.Printf("avg: %v\n", duration/time.Duration(count))
	fmt.Printf("total time: %v\n", timeToRun)
	fmt.Printf("req/sec: %v\n", float32(count)/(float32(timeToRun)/float32(time.Second)))
}

func randomRequest(client internalprotocol.ForecasterClient) (time.Duration, error) {

	latitude := (rand.Float32() * 180) - 90
	longitude := (rand.Float32() * 360) - 180
	request := internalprotocol.Location{
		Latitude:  latitude,
		Longitude: longitude,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	forecast, err := client.GetForecast(ctx, &request)
	end := time.Now()

	if err == nil {
		interpreted := internalprotocol.InterpretValues(forecast)
		ta, ok := interpreted["air_temperature_2m"]
		if !ok {
			log.Println("missing ta")
		} else {
			for i, v := range ta.Values {
				if v > 60 {
					log.Printf("%f/%f %v = %f", latitude, longitude, ta.Times[i], v)
				}
			}
		}
	}
	return end.Sub(start), err
}

func init() {
	rand.Seed(time.Now().Unix())
}
