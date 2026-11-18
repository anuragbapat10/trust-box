package repository

import (
	"context"
	"errors"
	"sync"
	"time"
)

type InMemoryRepository struct {
	mutex     sync.Mutex
	instances map[string]Instance
	capacity  int
}

const maxInstances = 1

func NewInMemoryRepository() *InMemoryRepository {
	repo, _ := NewInMemoryRepositoryWithCapacity(maxInstances)
	return repo
}

func NewInMemoryRepositoryWithCapacity(capacity int) (*InMemoryRepository, error) {
	if capacity < 1 {
		return nil, errors.New("repository: capacity must be >= 1")
	}
	return &InMemoryRepository{
		instances: make(map[string]Instance),
		capacity:  capacity,
	}, nil
}

func (r *InMemoryRepository) Create(ctx context.Context, inst Instance) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if inst.State == StateDeleted || inst.State == StateInactive {
		return ErrInvalidState
	}

	if len(r.instances) >= r.capacity {
		return ErrAtCapacity
	}

	r.instances[inst.ID] = inst
	return nil
}

func (r *InMemoryRepository) Get(ctx context.Context, id string) (Instance, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	inst, ok := r.instances[id]
	if !ok {
		return Instance{}, ErrNotFound
	}
	return inst, nil
}

func (r *InMemoryRepository) List(ctx context.Context) ([]Instance, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	out := make([]Instance, 0)
	for _, inst := range r.instances {
		if inst.State == StateDeleted {
			continue
		}
		out = append(out, inst)
	}
	return out, nil
}

func (r *InMemoryRepository) Delete(ctx context.Context, id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	inst, ok := r.instances[id]
	if !ok {
		return ErrNotFound
	}

	if inst.State == StateDeleted {
		return nil
	}

	inst.State = StateDeleted
	inst.DeleteTime = time.Now()
	r.instances[id] = inst
	return nil
}
