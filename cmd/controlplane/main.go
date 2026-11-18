// Command controlplane is the trust-box control plane: instance
// lifecycle, state metadata, and capacity enforcement.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/anuragbapat10/trust-box.git/cmd/controlplane/repository"
	pb "github.com/anuragbapat10/trust-box.git/internal/gen/trustbox/v1"
	"github.com/anuragbapat10/trust-box.git/internal/observability"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/credentials/insecure"
)

// TODO: port number and repository capacity from config

const serviceName = "controlplane"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	os.Exit(run(ctx))
}

func run(ctx context.Context) int {
	logger, cfg, mux, cleanup, err := observability.Setup(ctx, serviceName)
	if err != nil {
		slog.Error("observability setup failed", slog.String("error", err.Error()))
		return 1
	}
	defer func() {
		if err := cleanup(context.Background()); err != nil {
			logger.Error("observability shutdown failed", slog.String("error", err.Error()))
		}
	}()

	// gRPC server setup
	grpcAddr := os.Getenv("GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = ":50051"
	}

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Error("failed to listen on gRPC address", slog.String("addr", grpcAddr), slog.String("error", err.Error()))
		return 1
	}

	grpcServer := grpc.NewServer()
	repo := repository.NewInMemoryRepository()

	daemonAddr := os.Getenv("DAEMON_ADDR")
	if daemonAddr == "" {
		daemonAddr = "localhost:50052"
	}
	conn, err := grpc.NewClient(daemonAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("failed to dial daemon", slog.String("addr", daemonAddr), slog.String("error", err.Error()))
		return 1
	}
	daemonClient := pb.NewNodeDaemonClient(conn)

	pb.RegisterControlPlaneServer(grpcServer, &ControlPlaneServer{Repo: repo, DaemonClient: daemonClient})
	reflection.Register(grpcServer)

	// Run gRPC server in background
	grpcErrCh := make(chan error, 1)
	go func() {
		logger.Info("gRPC server listening", slog.String("addr", grpcAddr))
		if err := grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			grpcErrCh <- fmt.Errorf("gRPC server failed: %w", err)
		}
		close(grpcErrCh)
	}()

	// Graceful shutdown goroutine
	go func() {
		<-ctx.Done()
		logger.Info("shutting down gRPC server")
		grpcServer.GracefulStop()
	}()

	logger.Info("controlplane starting", slog.String("ops_addr", cfg.MetricsAddr), slog.String("grpc_addr", grpcAddr))

	if err := observability.Serve(ctx, cfg.MetricsAddr, mux); err != nil {
		logger.Error("ops server failed", slog.String("error", err.Error()))
		return 1
	}

	// Wait for gRPC server to finish
	if err := <-grpcErrCh; err != nil {
		logger.Error("gRPC server error", slog.String("error", err.Error()))
		return 1
	}

	return 0
}
