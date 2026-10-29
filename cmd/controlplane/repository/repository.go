// Package repository provides an abstraction layer and implementations for the datastore
package repository

import (
	"context"
	"errors"
	"time"
)

type InstanceRepository interface {
	Create(ctx context.Context, inst Instance) error
	Get(ctx context.Context, id string) (Instance, error)
	List(ctx context.Context) ([]Instance, error)
	Delete(ctx context.Context, id string) error
}

var (
	ErrNotFound     = errors.New("repository: not found")
	ErrAtCapacity   = errors.New("repository: at capacity")
	ErrInvalidState = errors.New("repository: invalid state")
)

const TombstoneTTL = 24 * time.Hour

var _ InstanceRepository = (*InMemoryRepository)(nil)
