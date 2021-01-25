package server

import (
	context "context"
	"errors"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"gitlab.met.no/forti/f2/internalprotocol"
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/health"
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/config"
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast"
)

func Run(conf *config.Configuration, port int) error {

	server, err := newGrpcServer(conf)
	if err != nil {
		return err
	}

	listenAddress := fmt.Sprintf(":%d", port)
	lis, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	internalprotocol.RegisterForecasterServer(s, server)
	grpc_health_v1.RegisterHealthServer(s, health.NewServer())

	log.Printf("listening on %s", listenAddress)

	return s.Serve(lis)
}

type grpcServer struct {
	internalprotocol.UnimplementedForecasterServer
	forecast *forecast.Forecast
}

func newGrpcServer(conf *config.Configuration) (*grpcServer, error) {
	forecast, err := forecast.New(conf)
	if err != nil {
		return nil, err
	}

	return &grpcServer{
		forecast: forecast,
	}, nil
}

func (s *grpcServer) GetForecast(ctx context.Context, in *internalprotocol.Location) (*internalprotocol.Forecast, error) {

	pointData, err := s.forecast.Get(in.Latitude, in.Longitude)
	if err != nil {
		if errors.Is(err, forecast.ErrOutsideAllGrids) {
			return &internalprotocol.Forecast{
				ForecastStatus: internalprotocol.ForecastStatus_OutsideAllGrids,
			}, nil
		}
		return nil, err
	}

	var size int
	for _, data := range pointData.Data {
		size += len(data.Data)
	}
	values := make([]float32, 0, size)
	parameterMeta := make([]*internalprotocol.ParameterMeta, 0, size)

	for _, data := range pointData.Data {
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
			UpdatedAt:  timestamppb.New(pointData.Meta.UpdatedAt),
			NextUpdate: timestamppb.New(pointData.Meta.NextUpdate),
		},
		ParameterMeta: parameterMeta,
		Data:          values,
	}

	return &forecast, nil
}
