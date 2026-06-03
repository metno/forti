package status

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/metno/forti/internalprotocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Status struct {
	LatUpdate time.Time `json:"last_update"`
}

type Fetcher struct {
	conn   *grpc.ClientConn
	client internalprotocol.ForecasterClient

	mutex      sync.RWMutex
	lastResult Status
	lastError  error
	lastCheck  time.Time
}

func NewFetcher(upstream string) (*Fetcher, error) {
	conn, err := grpc.Dial(upstream, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("could not connect to upstream: %w", err)
	}
	return &Fetcher{
		conn:   conn,
		client: internalprotocol.NewForecasterClient(conn),
	}, nil
}

func (f *Fetcher) Close() error {
	return f.conn.Close()
}

func (f *Fetcher) Get() (Status, error) {
	const checkInterval = 2 * time.Second
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	needUpdate := time.Now().After(f.lastCheck.Add(checkInterval))

	if needUpdate {
		f.mutex.RUnlock()

		f.mutex.Lock()
		now := time.Now()
		if now.After(f.lastCheck.Add(checkInterval)) {
			f.lastResult, f.lastError = getStatus(f.client)
			f.lastCheck = now
		}
		f.mutex.Unlock()

		f.mutex.RLock()
	}

	return f.lastResult, f.lastError
}

func getStatus(client internalprotocol.ForecasterClient) (Status, error) {
	log.Println("status")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	request := internalprotocol.GetForecastRequest{
		Latitude:  59,
		Longitude: 11,
		Altitude: &internalprotocol.Altitude{
			Value:    0,
			Override: true,
		},
	}

	forecast, err := client.GetForecast(ctx, &request)
	if err != nil {
		return Status{}, err
	}

	return Status{
		LatUpdate: forecast.ForecastMeta.UpdatedAt.AsTime().UTC().Round(time.Second),
	}, nil
}

func (f *Fetcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status, err := f.Get()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "internal server error")
		return
	}

	w.Header().Add("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(status)
}
