package main

import (
	"context"
	"io"
	"log/slog"
	"net"
	"os"
	"testing"

	pb "github.com/anuragbapat10/trust-box.git/internal/gen/trustbox/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

func setupTestDaemonServer(t *testing.T) (pb.NodeDaemonClient, func()) {
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	pb.RegisterNodeDaemonServer(s, NewNodeDaemonServer(logger))

	go func() {
		if err := s.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			t.Errorf("Server exited with error: %v", err)
		}
	}()

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	client := pb.NewNodeDaemonClient(conn)
	cleanup := func() {
		_ = conn.Close()
		s.Stop()
	}

	return client, cleanup
}

func TestNodeDaemonServer_StartInstance_EmptyID(t *testing.T) {
	client, cleanup := setupTestDaemonServer(t)
	defer cleanup()

	ctx := context.Background()
	_, err := client.StartInstance(ctx, &pb.StartInstanceRequest{Id: ""})
	if err == nil {
		t.Error("expected error when starting instance with empty ID, got nil")
	}
}


func TestNodeDaemonServer_StopInstance_EmptyID(t *testing.T) {
	client, cleanup := setupTestDaemonServer(t)
	defer cleanup()

	ctx := context.Background()
	_, err := client.StopInstance(ctx, &pb.StopInstanceRequest{Id: ""})
	if err == nil {
		t.Error("expected error when stopping instance with empty ID, got nil")
	}
}

func TestNodeDaemonServer_StopInstance_UnknownID_RemovesStaleSocket(t *testing.T) {
	client, cleanup := setupTestDaemonServer(t)
	defer cleanup()

	id := "nonexistent-test-vm"
	socketPath := "/tmp/firecracker-" + id + ".socket"

	if err := os.WriteFile(socketPath, []byte("stale"), 0o644); err != nil {
		t.Fatalf("create stale socket: %v", err)
	}
	defer os.Remove(socketPath)

	ctx := context.Background()
	_, err := client.StopInstance(ctx, &pb.StopInstanceRequest{Id: id})
	if err != nil {
		t.Fatalf("StopInstance failed: %v", err)
	}

	if _, err := os.Stat(socketPath); !os.IsNotExist(err) {
		t.Errorf("expected stale socket %s to be removed, got stat err %v", socketPath, err)
	}
}
