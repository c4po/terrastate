package storage

import (
	"context"

	"github.com/c4po/terrastate/internal/models"
)

type StateStorage interface {
	// State operations
	GetState(ctx context.Context, workspace, id string) (*models.State, error)
	PutState(ctx context.Context, state *models.State) error
	DeleteState(ctx context.Context, workspace, id string) error
	ListStates(ctx context.Context, workspace string) ([]models.State, error)

	// Lock operations
	Lock(ctx context.Context, lock *models.StateLock) error
	Unlock(ctx context.Context, workspace, id string) error
	GetLock(ctx context.Context, workspace, id string) (*models.StateLock, error)
}
