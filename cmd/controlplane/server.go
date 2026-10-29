package main

import (
	"context"
	"time"

	"github.com/anuragbapat10/trust-box.git/cmd/controlplane/repository"
	pb "github.com/anuragbapat10/trust-box.git/internal/gen/trustbox/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ControlPlaneServer struct {
	pb.UnimplementedControlPlaneServer
	Repo repository.InstanceRepository
}

func (s *ControlPlaneServer) CreateInstance(ctx context.Context, req *pb.CreateInstanceRequest) (*pb.Instance, error) {
	// TODO: This will be replaced later by daemon call
	now := time.Now()
	inst := repository.Instance{
		ID:         "inst-mock-1",
		State:      repository.StateActive,
		CreateTime: now,
	}
	// Until here

	if err := s.Repo.Create(ctx, inst); err != nil {
		return nil, err
	}

	return &pb.Instance{
		Id:         inst.ID,
		State:      pb.State_ACTIVE,
		CreateTime: timestamppb.New(inst.CreateTime), // TODO: This will be replaced by daemon generated timestamp
	}, nil
}

func (s *ControlPlaneServer) ListInstances(ctx context.Context, req *pb.ListInstancesRequest) (*pb.ListInstancesResponse, error) {
	insts, err := s.Repo.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*pb.Instance, 0, len(insts))
	for _, inst := range insts {
		state := pb.State_STATE_UNSPECIFIED
		switch inst.State {
		case repository.StateActive:
			state = pb.State_ACTIVE
		case repository.StateInactive:
			state = pb.State_INACTIVE
		case repository.StateDeleted:
			state = pb.State_DELETED
		}
		out = append(out, &pb.Instance{
			Id:         inst.ID,
			State:      state,
			CreateTime: timestamppb.New(inst.CreateTime),
			DeleteTime: func() *timestamppb.Timestamp {
				if inst.DeleteTime.IsZero() {
					return nil
				}
				return timestamppb.New(inst.DeleteTime)
			}(),
		})
	}
	return &pb.ListInstancesResponse{Instances: out}, nil
}

func (s *ControlPlaneServer) GetInstance(ctx context.Context, req *pb.GetInstanceRequest) (*pb.Instance, error) {
	inst, err := s.Repo.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	if inst.State == repository.StateDeleted {
		return nil, repository.ErrNotFound
	}
	state := pb.State_STATE_UNSPECIFIED
	switch inst.State {
	case repository.StateActive:
		state = pb.State_ACTIVE
	case repository.StateInactive:
		state = pb.State_INACTIVE
	case repository.StateDeleted:
		state = pb.State_DELETED
	}
	return &pb.Instance{
		Id:         inst.ID,
		State:      state,
		CreateTime: timestamppb.New(inst.CreateTime),
		DeleteTime: func() *timestamppb.Timestamp {
			if inst.DeleteTime.IsZero() {
				return nil
			}
			return timestamppb.New(inst.DeleteTime)
		}(),
	}, nil
}

func (s *ControlPlaneServer) DeleteInstance(ctx context.Context, req *pb.DeleteInstanceRequest) (*pb.DeleteInstanceResponse, error) {
	if err := s.Repo.Delete(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &pb.DeleteInstanceResponse{}, nil
}
