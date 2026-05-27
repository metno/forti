package server

import (
	"context"
	"fmt"
	"log"
	"math"
	"net"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"

	"gitlab.met.no/forti/f2/correctedforecaster/internal/correction"
	"gitlab.met.no/forti/f2/correctedforecaster/internal/health"
	"gitlab.met.no/forti/f2/correctedforecaster/internal/lookup"
	"gitlab.met.no/forti/f2/internalprotocol"
	"gitlab.met.no/forti/f2/parameters/radar"
)

func Run(upstream string, port int, topographyFiles []string) error {
	server, err := New(upstream, topographyFiles)
	if err != nil {
		return err
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)
	internalprotocol.RegisterForecasterServer(s, server)
	grpc_health_v1.RegisterHealthServer(s, health.NewServer(server.client))

	grpc_prometheus.EnableHandlingTimeHistogram()
	grpc_prometheus.Register(s)

	return s.Serve(lis)
}

type Server struct {
	internalprotocol.UnimplementedForecasterServer

	topo *lookup.Collection

	conn   *grpc.ClientConn
	client internalprotocol.ForecasterClient
}

func New(upstream string, topographyFiles []string) (*Server, error) {
	conn, err := grpc.Dial(upstream,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
	)
	if err != nil {
		return nil, fmt.Errorf("could not to upstream: %w", err)
	}

	client := internalprotocol.NewForecasterClient(conn)

	topo := lookup.NewCollection()
	if err := topo.Add(topographyFiles...); err != nil {
		return nil, fmt.Errorf("unable to add topography: %w", err)
	}

	return &Server{
		topo:   topo,
		conn:   conn,
		client: client,
	}, nil
}

func waitForUpstream(client internalprotocol.ForecasterClient) {
	var ctx context.Context
	request := internalprotocol.GetForecastRequest{
		Latitude:  60,
		Longitude: 10,
	}
	for {
		_, err := client.GetForecast(ctx, &request)
		if err == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}
}

func (s *Server) Close() error {
	return s.conn.Close()
}

func (s *Server) GetForecast(ctx context.Context, in *internalprotocol.GetForecastRequest) (*internalprotocol.Forecast, error) {
	forecast, err := s.client.GetForecast(ctx, in)
	if err != nil {
		log.Printf("unable to get forecast from rawdataforecaster: %s", err)
		return nil, fmt.Errorf("unable to get forecast from upstream: %w", err)
	}

	if forecast.ForecastStatus != internalprotocol.ForecastStatus_OK {
		return forecast, nil
	}

	if err := s.correct(in, forecast); err != nil {
		return nil, err
	}

	return forecast, nil
}

func (s *Server) correct(request *internalprotocol.GetForecastRequest, forecast *internalprotocol.Forecast) error {
	interpreted := internalprotocol.InterpretValues(forecast)
	if err := s.correctWithBetterTopography(request, interpreted); err != nil {
		return err
	}
	if err := s.correctWithRadarStatus(request, interpreted, forecast); err != nil {
		return err
	}
	return nil
}

func (s *Server) correctWithBetterTopography(request *internalprotocol.GetForecastRequest, interpreted map[string]internalprotocol.InterpretedData) error {
	modelAltitude, ok := getAltitude(interpreted)
	if !ok {
		return nil
	}

	var realAltitude float32
	if request.Altitude != nil && request.Altitude.Override {
		realAltitude = request.Altitude.Value
	} else {
		var err error
		realAltitude, err = s.topo.Lookup(float64(request.Latitude), float64(request.Longitude))
		if err != nil {
			if lookup.IsOutOfBounds(err) {
				realAltitude = *modelAltitude
			} else {
				log.Printf("unable to lookup topography for location: %s", err)
				return fmt.Errorf("unable to lookup topography for location: %w", err)
			}
		}
		realAltitude = float32(math.Round(float64(realAltitude)))
	}

	altitudeDiff := *modelAltitude - realAltitude
	if altitudeDiff < -100 || altitudeDiff > 100 {
		correction.UpdateTemperature(interpreted, altitudeDiff)
		correction.UpdateDewpointTemperature(interpreted)
		correction.UpdateSymbols(interpreted)
	}
	*modelAltitude = realAltitude

	return nil
}

func getAltitude(interpreted map[string]internalprotocol.InterpretedData) (*float32, bool) {
	altitude, ok := interpreted["altitude"]
	if !ok {
		return nil, false
	}
	return &altitude.Values[0], true
}

// correctWithRadarStatus removes all values for lwe_precipitation_rate if radar coverage is not ok.
func (s *Server) correctWithRadarStatus(request *internalprotocol.GetForecastRequest, interpreted map[string]internalprotocol.InterpretedData, forecast *internalprotocol.Forecast) error {
	if status, ok := interpreted["precipitation_status"]; ok {
		if radar.Coverage(status.Values[0]) != radar.OK {
			for _, m := range forecast.ParameterMeta {
				if m.Parameter == "lwe_precipitation_rate" {
					// This should cause all lookups to be of size 0
					m.Times = nil
					return nil
				}
			}
		}
	}
	return nil
}
