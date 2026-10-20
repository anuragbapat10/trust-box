package main

import (
	"context"

	pb "github.com/anuragbapat10/trust-box.git/internal/gen/trustbox/v1"
)

type ControlPlaneServer struct {
	pb.UnimplementedControlPlaneServer
}

func (s *ControlPlaneServer) CreateInstance(ctx context.Context, req *pb.CreateInstanceRequest) (*pb.Instance, error) {
	return nil, nil
}

func (s *ControlPlaneServer) ListInstances(ctx context.Context, req *pb.ListInstancesRequest) (*pb.ListInstancesResponse, error) {
	return nil, nil
}

func (s *ControlPlaneServer) GetInstance(ctx context.Context, req *pb.GetInstanceRequest) (*pb.Instance, error) {
	return nil, nil
}

func (s *ControlPlaneServer) DeleteInstance(ctx context.Context, req *pb.DeleteInstanceRequest) (*pb.DeleteInstanceResponse, error) {
	return nil, nil
}
