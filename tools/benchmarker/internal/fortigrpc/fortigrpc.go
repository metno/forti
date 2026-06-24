package fortigrpc

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/metno/forti/internal/internalprotocol"
	"github.com/metno/forti/tools/benchmarker/internal/location"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type FortiGRPC struct {
	conn   *grpc.ClientConn
	client internalprotocol.ForecasterClient
}

func NewClient(ctx context.Context, address string) (*FortiGRPC, error) {
	conn, err := grpc.DialContext(
		ctx, address,
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("did not connect: %w", err)
	}
	return &FortiGRPC{
		conn:   conn,
		client: internalprotocol.NewForecasterClient(conn),
	}, nil
}

func (f *FortiGRPC) Close() error {
	return f.conn.Close()
}

func (f *FortiGRPC) RandomRequest() (time.Duration, error) {
	l := location.Pick()
	request := internalprotocol.GetForecastRequest{
		Latitude:  l.Latitude,
		Longitude: l.Longitude,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	forecast, err := f.client.GetForecast(ctx, &request)
	end := time.Now()

	if err == nil {
		interpreted := internalprotocol.InterpretValues(forecast)
		ta, ok := interpreted["air_temperature_2m"]
		if !ok {
			log.Println("missing ta")
		} else {
			for i, v := range ta.Values {
				if v > 60 {
					log.Printf("%f/%f %v = %f", l.Latitude, l.Longitude, ta.Times[i], v)
				}
			}
		}
	}
	return end.Sub(start), err
}
