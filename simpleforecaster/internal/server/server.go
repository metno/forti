package server

import (
	context "context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"gitlab.met.no/forti/f2/simpleforecaster/internal/health"
	"gitlab.met.no/forti/f2/simpleforecaster/internal/server/forecast"
	"gitlab.met.no/forti/f2/simpleforecaster/pkg/forecaster"
)

type Configuration struct {
	Bucket string
	Groups []string
}

func Run(conf *Configuration) error {

	server, err := newGrpcServer(conf)
	if err != nil {
		return err
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	forecaster.RegisterForecasterServer(s, server)
	grpc_health_v1.RegisterHealthServer(s, health.NewServer())
	return s.Serve(lis)
}

type grpcServer struct {
	forecaster.UnimplementedForecasterServer
	forecaster *forecast.Forecast
}

func newGrpcServer(conf *Configuration) (*grpcServer, error) {
	forecast, err := forecast.New(conf.Bucket, conf.Groups)
	if err != nil {
		return nil, err
	}

	return &grpcServer{
		forecaster: forecast,
	}, nil
}

func (s *grpcServer) GetForecast(ctx context.Context, in *forecaster.Location) (*forecaster.Forecast, error) {

	pointData, err := s.forecaster.Get(in.Latitude, in.Longitude)
	if err != nil {
		return nil, err
	}

	var data []*forecaster.PointDataCollection
	for _, pd := range pointData.Data {

		var parameterMeta []*forecaster.ParameterMeta
		for parameter, meta := range pd.ParameterMeta {
			times := make([]*timestamppb.Timestamp, len(meta.Times))
			for i, t := range meta.Times {
				times[i] = timestamppb.New(t)
			}
			pm := &forecaster.ParameterMeta{
				Parameter: parameter,
				Units:     meta.Units,
				SliceFrom: int32(meta.SliceFrom),
				Times:     times,
			}
			parameterMeta = append(parameterMeta, pm)
		}

		collection := &forecaster.PointDataCollection{
			ParameterMeta: parameterMeta,
			Data:          pd.Data,
		}
		data = append(data, collection)
	}

	forecast := forecaster.Forecast{
		Meta: &forecaster.Meta{
			UpdatedAt:  timestamppb.New(pointData.Meta.UpdatedAt),
			NextUpdate: timestamppb.New(pointData.Meta.NextUpdate),
		},
		Data: data,
	}

	return &forecast, nil
}
