package disk

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/c4po/terrastate/internal/models"
)

type DiskStorage struct {
	basePath string
}

func NewDiskStorage(basePath string) *DiskStorage {
	return &DiskStorage{basePath: basePath}
}

func (d *DiskStorage) getStatePath(workspace, id string) string {
	return filepath.Join(d.basePath, workspace, id)
}

func (d *DiskStorage) getLockPath(workspace, id string) string {
	return filepath.Join(d.basePath, workspace, id+".lock")
}

func (d *DiskStorage) ensureDir(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0755)
}

func (d *DiskStorage) GetState(_ context.Context, workspace, id string) (*models.State, error) {
	path := d.getStatePath(workspace, id)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &models.State{
		ID:        id,
		Workspace: workspace,
		State:     data,
		UpdatedAt: fileInfo.ModTime(),
	}, nil
}

func (d *DiskStorage) PutState(_ context.Context, state *models.State) error {
	path := d.getStatePath(state.Workspace, state.ID)
	if err := d.ensureDir(path); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return os.WriteFile(path, state.State, 0644)
}

func (d *DiskStorage) DeleteState(_ context.Context, workspace, id string) error {
	return os.Remove(d.getStatePath(workspace, id))
}

func (d *DiskStorage) ListStates(_ context.Context, workspace string) ([]models.State, error) {
	dir := filepath.Join(d.basePath, workspace)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.State{}, nil
		}
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var states []models.State
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) == ".lock" {
			continue
		}

		fileInfo, err := entry.Info()
		if err != nil {
			continue
		}

		states = append(states, models.State{
			ID:        entry.Name(),
			Workspace: workspace,
			UpdatedAt: fileInfo.ModTime(),
		})
	}
	return states, nil
}

func (d *DiskStorage) Lock(_ context.Context, lock *models.StateLock) error {
	workspace, id := filepath.Split(lock.Path)
	path := d.getLockPath(workspace, id)

	if err := d.ensureDir(path); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.Marshal(lock)
	if err != nil {
		return fmt.Errorf("failed to marshal lock: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

func (d *DiskStorage) Unlock(_ context.Context, workspace, id string) error {
	return os.Remove(d.getLockPath(workspace, id))
}

func (d *DiskStorage) GetLock(_ context.Context, workspace, id string) (*models.StateLock, error) {
	data, err := os.ReadFile(d.getLockPath(workspace, id))
	if err != nil {
		return nil, fmt.Errorf("failed to read lock file: %w", err)
	}

	var lock models.StateLock
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("failed to unmarshal lock: %w", err)
	}

	return &lock, nil
}
