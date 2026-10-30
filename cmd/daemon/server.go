package main

import (
	"context"

	pb "github.com/anuragbapat10/trust-box.git/internal/gen/trustbox/v1"
)

type NodeDaemonServer struct {
	pb.UnimplementedNodeDaemonServer
}

func (s *NodeDaemonServer) StartInstance(ctx context.Context, req *pb.StartInstanceRequest) (*pb.NodeInstance, error) {
	return nil, nil
}

func (s *NodeDaemonServer) StopInstance(ctx context.Context, req *pb.StopInstanceRequest) (*pb.StopInstanceResponse, error) {
	return nil, nil
}
