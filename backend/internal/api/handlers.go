package api

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

// RegisterServices registers all gRPC services
func RegisterServices(server *grpc.Server, services *Services) {
	// TODO: Register actual gRPC services here
	// This is a placeholder for the gRPC service registration
	// In a real implementation, you would register your protobuf-generated services
}

// RegisterGatewayHandlers registers all gRPC-Gateway handlers
func RegisterGatewayHandlers(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	// TODO: Register actual gRPC-Gateway handlers here
	// This is a placeholder for the gRPC-Gateway handler registration
	// In a real implementation, you would register your protobuf-generated gateway handlers
	return nil
}
