package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	loanv1 "github.com/mrprofessor/loaner/gen/loan/v1"
	"github.com/mrprofessor/loaner/internal/config"
	"github.com/mrprofessor/loaner/internal/handler"
	"github.com/mrprofessor/loaner/internal/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg := config.Load()

	// Connect to the DB
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Verify database is reachable at startup rather than failing on first query
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("connected to database")

	st := store.New(pool)
	srv := handler.New(st)

	// gRPC server
	grpcServer := grpc.NewServer()
	loanv1.RegisterLoanServiceServer(grpcServer, srv)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	go func() {
		log.Printf("gRPC server listening on :%s", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC error: %v", err)
		}
	}()

	// HTTP gateway
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if err := loanv1.RegisterLoanServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%s", cfg.GRPCPort), opts); err != nil {
		log.Fatalf("failed to register gateway: %v", err)
	}

	httpServer := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: mux}
	go func() {
		log.Printf("HTTP gateway listening on :%s", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")
	grpcServer.GracefulStop()
	httpServer.Shutdown(context.Background())
}
