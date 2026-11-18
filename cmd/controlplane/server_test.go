package main

import (
	"context"
	"net"
	"testing"

	"github.com/anuragbapat10/trust-box.git/cmd/controlplane/repository"
	pb "github.com/anuragbapat10/trust-box.git/internal/gen/trustbox/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

func setupTestServer(t *testing.T) (pb.ControlPlaneClient, *repository.InMemoryRepository, func()) {
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	repo := repository.NewInMemoryRepository()
	pb.RegisterControlPlaneServer(s, &ControlPlaneServer{Repo: repo})

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

	client := pb.NewControlPlaneClient(conn)
	cleanup := func() {
		_ = conn.Close()
		s.Stop()
	}

	return client, repo, cleanup
}

func TestControlPlaneServer_CreateInstance(t *testing.T) {
	client, repo, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	req := &pb.CreateInstanceRequest{}

	resp, err := client.CreateInstance(ctx, req)
	if err != nil {
		t.Fatalf("CreateInstance failed: %v", err)
	}

	if resp.GetId() == "" {
		t.Error("expected non-empty instance ID in response")
	}
	if resp.GetState() != pb.State_ACTIVE {
		t.Errorf("expected state ACTIVE, got %v", resp.GetState())
	}
	if resp.GetCreateTime() == nil {
		t.Error("expected non-nil CreateTime in response")
	}

	// Verify instance was stored in repository
	inst, err := repo.Get(ctx, resp.GetId())
	if err != nil {
		t.Fatalf("instance not found in repository: %v", err)
	}
	if inst.ID != resp.GetId() {
		t.Errorf("expected repo instance ID %s, got %s", resp.GetId(), inst.ID)
	}
}

func TestControlPlaneServer_CreateInstance_AtCapacity(t *testing.T) {
	client, _, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	req := &pb.CreateInstanceRequest{}

	// First creation succeeds (uses capacity 1)
	_, err := client.CreateInstance(ctx, req)
	if err != nil {
		t.Fatalf("first CreateInstance failed: %v", err)
	}

	// Second creation should fail due to maxInstances=1 limit
	_, err = client.CreateInstance(ctx, req)
	if err == nil {
		t.Error("expected error on second CreateInstance at capacity, got nil")
	}
}

func TestControlPlaneServer_GetInstance(t *testing.T) {
	client, _, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	created, err := client.CreateInstance(ctx, &pb.CreateInstanceRequest{})
	if err != nil {
		t.Fatalf("CreateInstance failed: %v", err)
	}

	got, err := client.GetInstance(ctx, &pb.GetInstanceRequest{Id: created.GetId()})
	if err != nil {
		t.Fatalf("GetInstance failed: %v", err)
	}

	if got.GetId() != created.GetId() {
		t.Errorf("expected ID %s, got %s", created.GetId(), got.GetId())
	}
	if got.GetState() != pb.State_ACTIVE {
		t.Errorf("expected state ACTIVE, got %v", got.GetState())
	}
	if got.GetCreateTime() == nil {
		t.Error("expected non-nil CreateTime")
	}
}

func TestControlPlaneServer_GetInstance_NotFound(t *testing.T) {
	client, _, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	_, err := client.GetInstance(ctx, &pb.GetInstanceRequest{Id: "non-existent-id"})
	if err == nil {
		t.Error("expected error for non-existent instance ID, got nil")
	}
}

func TestControlPlaneServer_ListInstances_Empty(t *testing.T) {
	client, _, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	resp, err := client.ListInstances(ctx, &pb.ListInstancesRequest{})
	if err != nil {
		t.Fatalf("ListInstances failed: %v", err)
	}

	if len(resp.GetInstances()) != 0 {
		t.Errorf("expected 0 instances, got %d", len(resp.GetInstances()))
	}
}

func TestControlPlaneServer_ListInstances_OneInstance(t *testing.T) {
	client, _, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	created, err := client.CreateInstance(ctx, &pb.CreateInstanceRequest{})
	if err != nil {
		t.Fatalf("CreateInstance failed: %v", err)
	}

	resp, err := client.ListInstances(ctx, &pb.ListInstancesRequest{})
	if err != nil {
		t.Fatalf("ListInstances failed: %v", err)
	}

	if len(resp.GetInstances()) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(resp.GetInstances()))
	}

	item := resp.GetInstances()[0]
	if item.GetId() != created.GetId() {
		t.Errorf("expected ID %s, got %s", created.GetId(), item.GetId())
	}
	if item.GetState() != pb.State_ACTIVE {
		t.Errorf("expected state ACTIVE, got %v", item.GetState())
	}
	if item.GetCreateTime() == nil {
		t.Error("expected non-nil CreateTime")
	}
}

func TestControlPlaneServer_DeleteInstance(t *testing.T) {
	client, _, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	created, err := client.CreateInstance(ctx, &pb.CreateInstanceRequest{})
	if err != nil {
		t.Fatalf("CreateInstance failed: %v", err)
	}

	_, err = client.DeleteInstance(ctx, &pb.DeleteInstanceRequest{Id: created.GetId()})
	if err != nil {
		t.Fatalf("DeleteInstance failed: %v", err)
	}

	// Verify GetInstance now returns error for deleted instance
	_, err = client.GetInstance(ctx, &pb.GetInstanceRequest{Id: created.GetId()})
	if err == nil {
		t.Error("expected GetInstance to fail for deleted instance, got nil error")
	}

	// Verify ListInstances excludes deleted instance
	listResp, err := client.ListInstances(ctx, &pb.ListInstancesRequest{})
	if err != nil {
		t.Fatalf("ListInstances failed: %v", err)
	}
	if len(listResp.GetInstances()) != 0 {
		t.Errorf("expected 0 active instances after deletion, got %d", len(listResp.GetInstances()))
	}
}

func TestControlPlaneServer_DeleteInstance_NotFound(t *testing.T) {
	client, _, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	_, err := client.DeleteInstance(ctx, &pb.DeleteInstanceRequest{Id: "non-existent-id"})
	if err == nil {
		t.Error("expected error for deleting non-existent ID, got nil")
	}
}

func TestControlPlaneServer_DeleteInstance_Idempotent(t *testing.T) {
	client, _, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	created, err := client.CreateInstance(ctx, &pb.CreateInstanceRequest{})
	if err != nil {
		t.Fatalf("CreateInstance failed: %v", err)
	}

	// First delete
	_, err = client.DeleteInstance(ctx, &pb.DeleteInstanceRequest{Id: created.GetId()})
	if err != nil {
		t.Fatalf("first DeleteInstance failed: %v", err)
	}

	// Second delete on same ID should succeed (idempotent)
	_, err = client.DeleteInstance(ctx, &pb.DeleteInstanceRequest{Id: created.GetId()})
	if err != nil {
		t.Errorf("second DeleteInstance (idempotent) failed: %v", err)
	}
}


