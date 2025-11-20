package health

import (
	"context"

	"google.golang.org/grpc/health/grpc_health_v1"
)

// Server implements grpc's health check protocol: https://github.com/grpc/grpc/blob/v1.15.0/doc/health-checking.md
type Server struct {
	grpc_health_v1.UnimplementedHealthServer // https://pkg.go.dev/google.golang.org/grpc@v1.77.0/health/grpc_health_v1#UnimplementedHealthServer
}

// NewServer creates a new server.
func NewServer() *Server {
	return &Server{}
}

// Check reports on the health of the system.
func (h *Server) Check(ctx context.Context, request *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	status := grpc_health_v1.HealthCheckResponse_SERVING
	if request.Service != "" {
		status = grpc_health_v1.HealthCheckResponse_SERVICE_UNKNOWN
	}
	return &grpc_health_v1.HealthCheckResponse{Status: status}, nil
}

// Watch sends notifications about health changes
func (h *Server) Watch(request *grpc_health_v1.HealthCheckRequest, server grpc_health_v1.Health_WatchServer) error {
	response := &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING}
	server.Send(response)
	return nil
}
