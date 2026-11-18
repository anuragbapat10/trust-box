package repository

import "time"

type State int

const (
	StateUnspecified State = iota
	StateActive
	StateInactive
	StateDeleted
)

type Instance struct {
	ID         string
	State      State
	CreateTime time.Time
	DeleteTime time.Time
}
