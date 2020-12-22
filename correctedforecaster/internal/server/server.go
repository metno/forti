package server

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"

	"gitlab.met.no/forti/f2/simpleforecaster/pkg/forecaster"

	"gitlab.met.no/forti/f2/correctedforecaster/internal/correction"
	"gitlab.met.no/forti/f2/correctedforecaster/internal/health"
	"gitlab.met.no/forti/f2/correctedforecaster/internal/lookup"
)

func Run(upstream string, topographyFiles []string) error {
	server, err := New(upstream, topographyFiles)
	if err != nil {
		return err
	}

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	forecaster.RegisterForecasterServer(s, server)
	grpc_health_v1.RegisterHealthServer(s, health.NewServer(server.client))
	return s.Serve(lis)
}

type Server struct {
	forecaster.UnimplementedForecasterServer

	topo *lookup.Collection

	conn   *grpc.ClientConn
	client forecaster.ForecasterClient
}

func New(upstream string, topographyFiles []string) (*Server, error) {
	conn, err := grpc.Dial(upstream, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("could not to upstream: %w", err)
	}

	client := forecaster.NewForecasterClient(conn)

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

func waitForUpstream(client forecaster.ForecasterClient) {
	var ctx context.Context
	request := forecaster.Location{
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

func (s *Server) GetForecast(ctx context.Context, in *forecaster.Location) (*forecaster.Forecast, error) {
	forecast, err := s.client.GetForecast(ctx, in)
	if err != nil {
		return nil, fmt.Errorf("unable to get forecast from upstream: %w", err)
	}

	if err := s.correct(in, forecast); err != nil {
		return nil, err
	}

	return forecast, nil
}

func (s *Server) correct(request *forecaster.Location, forecast *forecaster.Forecast) error {

	interpreted := forecaster.InterpretValues(forecast)

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
				return fmt.Errorf("unable to lookup topography for location: %w", err)
			}
		}
	}

	altitudeDiff := *modelAltitude - realAltitude
	if altitudeDiff < -100 || altitudeDiff > 100 {
		correction.UpdateTemperature(interpreted, int(altitudeDiff))
		correction.UpdateDewpointTemperature(interpreted)
		correction.UpdateSymbols(interpreted)
		correction.UpdateSymbols6h(interpreted)
	}
	*modelAltitude = realAltitude

	return nil
}

func getAltitude(interpreted map[string]forecaster.InterpretedData) (*float32, bool) {
	altitude, ok := interpreted["altitude"]
	if !ok {
		return nil, false
	}
	return &altitude.Values[0], true
}
