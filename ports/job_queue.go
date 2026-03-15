package ports

import (
	"context"

	"github.com/sakibalam/bloodconnect/domain"
)

// JobQueue defines the interface for interacting with background tasks
type JobQueue interface {
	Enqueue(ctx context.Context, job *domain.Job) error
	FetchNextAvailable(ctx context.Context) (*domain.Job, error)
	MarkStatus(ctx context.Context, id string, status domain.JobStatus) error
}
