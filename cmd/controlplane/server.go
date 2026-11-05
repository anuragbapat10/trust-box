package main

import (
	"context"
	"time"

	"github.com/anuragbapat10/trust-box.git/cmd/controlplane/repository"
	pb "github.com/anuragbapat10/trust-box.git/internal/gen/trustbox/v1"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ControlPlaneServer struct {
	pb.UnimplementedControlPlaneServer
	Repo         repository.InstanceRepository
	DaemonClient pb.NodeDaemonClient
}

func (s *ControlPlaneServer) CreateInstance(ctx context.Context, req *pb.CreateInstanceRequest) (*pb.Instance, error) {
	id := uuid.NewString()

	var createTime time.Time
	if s.DaemonClient != nil {
		resp, err := s.DaemonClient.StartInstance(ctx, &pb.StartInstanceRequest{Id: id})
		if err != nil {
			return nil, err
		}
		createTime = resp.GetCreateTime().AsTime()
	} else {
		createTime = time.Now()
	}

	inst := repository.Instance{
		ID:         id,
		State:      repository.StateActive,
		CreateTime: createTime,
	}

	if err := s.Repo.Create(ctx, inst); err != nil {
		return nil, err
	}

	return &pb.Instance{
		Id:         inst.ID,
		State:      pb.State_ACTIVE,
		CreateTime: timestamppb.New(inst.CreateTime),
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
