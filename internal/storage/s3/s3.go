package s3

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/c4po/terrastate/internal/models"
)

type S3Storage struct {
	client     *s3.Client
	bucketName string
	prefix     string // Optional prefix for all keys
}

func NewS3Storage(client *s3.Client, bucketName string) *S3Storage {
	prefix := os.Getenv("S3_PREFIX")
	return &S3Storage{
		client:     client,
		bucketName: bucketName,
		prefix:     prefix,
	}
}

// getFullKey returns the complete S3 key including any configured prefix
func (s *S3Storage) getFullKey(workspace, id string) string {
	key := fmt.Sprintf("%s/%s", workspace, id)
	if s.prefix != "" {
		key = fmt.Sprintf("%s/%s", s.prefix, key)
	}
	return key
}

func (s *S3Storage) GetState(ctx context.Context, workspace, id string) (*models.State, error) {
	key := s.getFullKey(workspace, id)

	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get state from S3: %w", err)
	}
	defer output.Body.Close()

	state := &models.State{
		ID:        id,
		Workspace: workspace,
	}

	// Read the state data
	stateData, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read state data: %w", err)
	}
	state.State = stateData

	// Set metadata from object
	if output.LastModified != nil {
		state.UpdatedAt = *output.LastModified
	}
	if etag := output.ETag; etag != nil {
		state.MD5 = *etag
	}

	return state, nil
}

func (s *S3Storage) DeleteState(ctx context.Context, workspace, id string) error {
	key := s.getFullKey(workspace, id)
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete state from S3: %w", err)
	}
	return nil
}

func (s *S3Storage) GetLock(ctx context.Context, workspace, id string) (*models.StateLock, error) {
	key := s.getFullKey(workspace, id) + ".lock"
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get lock from S3: %w", err)
	}
	defer output.Body.Close()

	var lock models.StateLock
	if err := json.NewDecoder(output.Body).Decode(&lock); err != nil {
		return nil, fmt.Errorf("failed to decode lock data: %w", err)
	}
	return &lock, nil
}

func (s *S3Storage) ListStates(ctx context.Context, workspace string) ([]models.State, error) {
	prefix := s.getFullKey(workspace, "")
	output, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucketName),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list states from S3: %w", err)
	}

	var states []models.State
	for _, obj := range output.Contents {
		states = append(states, models.State{
			ID:        *obj.Key,
			Workspace: workspace,
			UpdatedAt: *obj.LastModified,
		})
	}
	return states, nil
}

func (s *S3Storage) Lock(ctx context.Context, lock *models.StateLock) error {
	key := s.getFullKey(lock.Path, "") + ".lock"
	data, err := json.Marshal(lock)
	if err != nil {
		return fmt.Errorf("failed to marshal lock data: %w", err)
	}

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	return err
}

func (s *S3Storage) PutState(ctx context.Context, state *models.State) error {
	key := s.getFullKey(state.Workspace, state.ID)
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader(state.State),
	})
	return err
}

func (s *S3Storage) Unlock(ctx context.Context, workspace, id string) error {
	key := s.getFullKey(workspace, id) + ".lock"
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	return err
}

// Implement other interface methods...
