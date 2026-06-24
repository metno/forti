package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/metno/forti/tools/benchmarker/internal/fortigrpc"
	"github.com/metno/forti/tools/benchmarker/internal/fortihttp"
)

func main() {
	connectionType := flag.String("type", "grpc", "Type of connection (http or grpc)")
	address := flag.String("address", "localhost:50051", "Server to connect to")
	threads := flag.Int("threads", 10, "number of threads")
	httpReadDelay := flag.Int("http-read-delay-ms", 0, "Wait the given number of milliseconds before reading http response")
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

	client, err := getRequester(*connectionType, *address, *httpReadDelay)
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()

	start := time.Now()
	var wait sync.WaitGroup
	for i := 0; i < *threads; i++ {
		wait.Add(1)
		go func() {
			for !stop {
				duration, err := client.RandomRequest()
				if err != nil {
					log.Println(err)
					errCh <- err
				}
				timesCh <- duration
			}
			wait.Done()
		}()
	}

	fmt.Println("Running. Press CTRL-C to exit")

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

func getRequester(connectionType, address string, httpReadDelay int) (RandomRequester, error) {
	switch connectionType {
	case "grpc":
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return fortigrpc.NewClient(ctx, address)
	case "http":
		return fortihttp.NewClient(address, time.Duration(httpReadDelay)*time.Millisecond)
	}
	return nil, fmt.Errorf("invalid connection type: %s", connectionType)
}

type RandomRequester interface {
	io.Closer
	RandomRequest() (time.Duration, error)
}
