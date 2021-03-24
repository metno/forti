package server

import (
	context "context"
	"errors"
	"fmt"
	"log"
	"net"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"gitlab.met.no/forti/f2/internalprotocol"
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/health"
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/config"
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast"
)

func Run(conf *config.Configuration, port int) error {

	server, err := newForecastServer(conf)
	if err != nil {
		return err
	}

	listenAddress := fmt.Sprintf(":%d", port)
	lis, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return err
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)
	internalprotocol.RegisterForecasterServer(s, server)
	grpc_health_v1.RegisterHealthServer(s, health.NewServer())

	grpc_prometheus.EnableHandlingTimeHistogram()
	grpc_prometheus.Register(s)

	log.Printf("listening on %s", listenAddress)

	return s.Serve(lis)
}

type forecastServer struct {
	internalprotocol.UnimplementedForecasterServer
	forecast *forecast.Forecast
}

func newForecastServer(conf *config.Configuration) (*forecastServer, error) {
	forecast, err := forecast.New(conf)
	if err != nil {
		return nil, err
	}

	return &forecastServer{
		forecast: forecast,
	}, nil
}

func (s *forecastServer) GetForecast(ctx context.Context, in *internalprotocol.GetForecastRequest) (*internalprotocol.Forecast, error) {

	locationData, err := s.forecast.Get(in.Latitude, in.Longitude)
	if err != nil {
		if errors.Is(err, forecast.ErrOutsideAllGrids) {
			return &internalprotocol.Forecast{
				ForecastStatus: internalprotocol.ForecastStatus_OutsideAllGrids,
			}, nil
		}
		return nil, err
	}

	var dataSize int
	var parameterCount int
	for _, data := range locationData.Data {
		dataSize += len(data.Data)
		parameterCount += len(data.ParameterMeta)
	}
	values := make([]float32, 0, dataSize)
	gridLocation := internalprotocol.Location{
		Latitude:  locationData.GridLocation.Lat,
		Longitude: locationData.GridLocation.Long,
	}

	parameterMeta := make([]*internalprotocol.ParameterMeta, 0, parameterCount)

	for _, data := range locationData.Data {
		for parameter, meta := range data.ParameterMeta {
			times := make([]*timestamppb.Timestamp, len(meta.Times))
			for i, t := range meta.Times {
				times[i] = timestamppb.New(t)
			}
			pm := &internalprotocol.ParameterMeta{
				Parameter: parameter,
				Units:     meta.Units,
				SliceFrom: int32(meta.SliceFrom + len(values)),
				Times:     times,
			}
			parameterMeta = append(parameterMeta, pm)
		}

		values = append(values, data.Data...)
	}

	forecast := internalprotocol.Forecast{
		ForecastStatus: internalprotocol.ForecastStatus_OK,
		ForecastMeta: &internalprotocol.ForecastMeta{
			UpdatedAt:    timestamppb.New(locationData.Meta.UpdatedAt),
			NextUpdate:   timestamppb.New(locationData.Meta.NextUpdate),
			GridLocation: &gridLocation,
		},
		ParameterMeta: parameterMeta,
		Data:          values,
	}

	return &forecast, nil
}
