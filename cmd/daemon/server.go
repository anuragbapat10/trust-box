package main

import (
	"context"
	"fmt"
	"log/slog"

	pb "github.com/anuragbapat10/trust-box.git/internal/gen/trustbox/v1"
)

type NodeDaemonServer struct {
	pb.UnimplementedNodeDaemonServer
	logger         *slog.Logger
	instanceServer *InstanceServer
}

func NewNodeDaemonServer(logger *slog.Logger) *NodeDaemonServer {
	return &NodeDaemonServer{
		logger:         logger,
		instanceServer: NewInstanceServer(logger),
	}
}

func (s *NodeDaemonServer) StartInstance(ctx context.Context, req *pb.StartInstanceRequest) (*pb.NodeInstance, error) {
	return s.instanceServer.startup(ctx, req.GetId())
}

func (s *NodeDaemonServer) StopInstance(ctx context.Context, req *pb.StopInstanceRequest) (*pb.StopInstanceResponse, error) {
	if req.GetId() == "" {
		return nil, fmt.Errorf("instance id cannot be empty")
	}
	if err := s.instanceServer.teardown(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &pb.StopInstanceResponse{}, nil
}
