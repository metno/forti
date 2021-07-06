package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"sync"
	"time"
)

func main() {
	address := flag.String("address", "localhost:8080/api/forecast/v2/complete", "Server to connect to")
	threads := flag.Int("threads", 10, "number of threads")
	flag.Parse()

	timesCh := make(chan time.Duration)

	errCh := make(chan error)
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

	start := time.Now()
	var wait sync.WaitGroup
	for i := 0; i < *threads; i++ {
		wait.Add(1)
		go func() {
			for !stop {
				duration, err := randomRequest(*address)
				if err != nil {
					errCh <- err
				}
				timesCh <- duration
			}
			wait.Done()
		}()
	}
	go func() {
		wait.Wait()
		close(timesCh)
	}()

	fmt.Println("running. Press CTRL-C to exit")

	statsCollector := newTimeReporter()
	for t := range timesCh {
		statsCollector.Report(t)
	}

	end := time.Now()

	timeToRunInSeconds := end.Sub(start) / time.Second

	fmt.Println()
	fmt.Printf("threads: %d\n", *threads)
	fmt.Printf("count: %d\n", statsCollector.Count())
	fmt.Printf("errors: %v\n", errors)
	fmt.Printf("avg: %v\n", statsCollector.Avg())
	fmt.Printf("min: %v\n", statsCollector.Quantile(0))
	fmt.Printf("median: %v\n", statsCollector.Quantile(.5))
	fmt.Printf("90th percentile: %v\n", statsCollector.Quantile(.9))
	fmt.Printf("99th percentile: %v\n", statsCollector.Quantile(.99))
	fmt.Printf("max: %v\n", statsCollector.Quantile(1))
	fmt.Printf("total time (all threads): %v\n", statsCollector.TotalTime())
	fmt.Printf("req/sec: %v\n", float32(statsCollector.Count())/(float32(timeToRunInSeconds)))
}

type timeReporter struct {
	times      map[time.Duration]int
	bucketSize time.Duration
	totalTime  time.Duration
	count      int
}

func newTimeReporter() *timeReporter {
	return &timeReporter{
		times:      make(map[time.Duration]int),
		bucketSize: time.Millisecond,
	}
}

func (r *timeReporter) Report(d time.Duration) {
	r.totalTime += d
	r.count++
	bucket := d.Round(r.bucketSize)
	r.times[bucket]++
}

func (r *timeReporter) Count() int {
	return r.count
}

func (r *timeReporter) Avg() time.Duration {
	return r.totalTime / time.Duration(r.count)
}

func (r *timeReporter) TotalTime() time.Duration {
	return r.totalTime
}

func (r *timeReporter) AvgPerSecond() float32 {
	return float32(r.count) / (float32(r.totalTime) / float32(time.Second))
}

func (r *timeReporter) Quantile(q float32) time.Duration {
	var totalCount int
	sortedBuckets := make([]int, 0, len(r.times))
	for bucket, count := range r.times {
		sortedBuckets = append(sortedBuckets, int(bucket))
		totalCount += count
	}
	sort.Ints(sortedBuckets)

	targetCount := float32(totalCount) * q
	var currentCount int
	for _, b := range sortedBuckets {
		currentCount += r.times[time.Duration(b)]
		if targetCount < float32(currentCount) {
			return time.Duration(b)
		}
	}
	lastBucket := sortedBuckets[len(sortedBuckets)-1]
	return time.Duration(lastBucket)
}

func randomRequest(address string) (time.Duration, error) {

	var latitude, longitude float32
	if rand.Float32() > 0.5 {
		// around Norway
		longitude = (rand.Float32() * 5) + 4
		latitude = (rand.Float32() * 10) + 59

	} else {
		// entire world
		latitude = (rand.Float32() * 180) - 90
		longitude = (rand.Float32() * 360) - 180
	}

	url := fmt.Sprintf("%s?lat=%.4f&lon=%.4f", address, latitude, longitude)

	start := time.Now()
	response, err := http.Get(url)
	if err == nil {
		response.Body.Close()
	}
	end := time.Now()

	// if err == nil {
	// 	interpreted := internalprotocol.InterpretValues(forecast)
	// 	ta, ok := interpreted["air_temperature_2m"]
	// 	if !ok {
	// 		log.Println("missing ta")
	// 	} else {
	// 		for i, v := range ta.Values {
	// 			if v > 60 {
	// 				log.Printf("%f/%f %v = %f", latitude, longitude, ta.Times[i], v)
	// 			}
	// 		}
	// 	}
	// }
	return end.Sub(start), err
}

func init() {
	rand.Seed(time.Now().Unix())
}
