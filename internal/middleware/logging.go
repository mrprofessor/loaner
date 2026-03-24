package middleware

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor is a gRPC unary interceptor that logs the method, duration, and status of each request.
func LoggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	slog.Info("request",
		"method", info.FullMethod,
		"duration", time.Since(start).String(),
		"status", status.Code(err).String(),
	)
	return resp, err
}
