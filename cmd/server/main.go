package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	loanv1 "github.com/mrprofessor/loaner/gen/loan/v1"
	"github.com/mrprofessor/loaner/internal/handler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	grpcPort := env("GRPC_PORT", "9090")
	httpPort := env("HTTP_PORT", "8080")

	srv := handler.New()

	// gRPC server
	grpcServer := grpc.NewServer()
	loanv1.RegisterLoanServiceServer(grpcServer, srv)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	go func() {
		log.Printf("gRPC server listening on :%s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC error: %v", err)
		}
	}()

	// HTTP gateway
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if err := loanv1.RegisterLoanServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%s", grpcPort), opts); err != nil {
		log.Fatalf("failed to register gateway: %v", err)
	}

	httpServer := &http.Server{Addr: ":" + httpPort, Handler: mux}
	go func() {
		log.Printf("HTTP gateway listening on :%s", httpPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")
	grpcServer.GracefulStop()
	httpServer.Shutdown(context.Background())
}
