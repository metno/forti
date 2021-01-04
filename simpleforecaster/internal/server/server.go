package server

import (
	context "context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"gitlab.met.no/forti/f2/internalprotocol"
	"gitlab.met.no/forti/f2/simpleforecaster/internal/health"
	"gitlab.met.no/forti/f2/simpleforecaster/internal/server/forecast"
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
	internalprotocol.RegisterForecasterServer(s, server)
	grpc_health_v1.RegisterHealthServer(s, health.NewServer())
	return s.Serve(lis)
}

type grpcServer struct {
	internalprotocol.UnimplementedForecasterServer
	internalprotocol *forecast.Forecast
}

func newGrpcServer(conf *Configuration) (*grpcServer, error) {
	forecast, err := forecast.New(conf.Bucket, conf.Groups)
	if err != nil {
		return nil, err
	}

	return &grpcServer{
		internalprotocol: forecast,
	}, nil
}

func (s *grpcServer) GetForecast(ctx context.Context, in *internalprotocol.Location) (*internalprotocol.Forecast, error) {

	pointData, err := s.internalprotocol.Get(in.Latitude, in.Longitude)
	if err != nil {
		return nil, err
	}

	var data []*internalprotocol.PointDataCollection
	for _, pd := range pointData.Data {

		var parameterMeta []*internalprotocol.ParameterMeta
		for parameter, meta := range pd.ParameterMeta {
			times := make([]*timestamppb.Timestamp, len(meta.Times))
			for i, t := range meta.Times {
				times[i] = timestamppb.New(t)
			}
			pm := &internalprotocol.ParameterMeta{
				Parameter: parameter,
				Units:     meta.Units,
				SliceFrom: int32(meta.SliceFrom),
				Times:     times,
			}
			parameterMeta = append(parameterMeta, pm)
		}

		collection := &internalprotocol.PointDataCollection{
			ParameterMeta: parameterMeta,
			Data:          pd.Data,
		}
		data = append(data, collection)
	}

	forecast := internalprotocol.Forecast{
		Meta: &internalprotocol.Meta{
			UpdatedAt:  timestamppb.New(pointData.Meta.UpdatedAt),
			NextUpdate: timestamppb.New(pointData.Meta.NextUpdate),
		},
		Data: data,
	}

	return &forecast, nil
}
