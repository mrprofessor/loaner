package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	loanv1 "github.com/mrprofessor/loaner/gen/loan/v1"
	"github.com/mrprofessor/loaner/internal/config"
	"github.com/mrprofessor/loaner/internal/handler"
	"github.com/mrprofessor/loaner/internal/middleware"
	"github.com/mrprofessor/loaner/internal/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg := config.Load()

	// Connect to the DB
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		slog.Error("failed to ping database", "error", err)
		os.Exit(1)
	}
	slog.Info("connected to database")

	st := store.New(pool)
	srv := handler.New(st)

	// gRPC server
	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		middleware.RecoveryInterceptor,
		middleware.LoggingInterceptor,
	))
	loanv1.RegisterLoanServiceServer(grpcServer, srv)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		slog.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	go func() {
		slog.Info("gRPC server listening", "port", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("gRPC error", "error", err)
			os.Exit(1)
		}
	}()

	// HTTP gateway
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if err := loanv1.RegisterLoanServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%s", cfg.GRPCPort), opts); err != nil {
		slog.Error("failed to register gateway", "error", err)
		os.Exit(1)
	}

	httpServer := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: mux}
	go func() {
		slog.Info("HTTP gateway listening", "port", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down...")
	grpcServer.GracefulStop()
	httpServer.Shutdown(context.Background())
}
