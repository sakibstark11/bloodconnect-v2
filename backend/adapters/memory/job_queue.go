package memory

import (
	"context"
	"sync"
	"time"

	"bloodconnect/application"
	"bloodconnect/application/domain"
)

type InMemoryJobQueue struct {
	mu   sync.Mutex
	jobs []*domain.Job
}

func NewJobQueue() application.JobQueue {
	return &InMemoryJobQueue{
		jobs: make([]*domain.Job, 0),
	}
}

func (q *InMemoryJobQueue) Enqueue(ctx context.Context, job *domain.Job) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.jobs = append(q.jobs, job)
	return nil
}

func (q *InMemoryJobQueue) FetchNextAvailable(ctx context.Context) (*domain.Job, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now().UTC()
	var nextJob *domain.Job
	var earliestRunAt time.Time

	for _, job := range q.jobs {
		if job.Status == domain.JobStatusPending {
			runAt, _ := time.Parse("2006-01-02T15:04:05.000Z", string(job.RunAt))
			if runAt.Before(now) {
				if nextJob == nil || runAt.Before(earliestRunAt) {
					nextJob = job
					earliestRunAt = runAt
				}
			}
		}
	}
	return nextJob, nil
}

func (q *InMemoryJobQueue) MarkStatus(ctx context.Context, id domain.JobID, status domain.JobStatus) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, job := range q.jobs {
		if job.ID == id {
			job.Status = status
			job.UpdatedAt = domain.Now()
			return nil
		}
	}
	return nil
}

func (q *InMemoryJobQueue) Delete(ctx context.Context, id domain.JobID) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, job := range q.jobs {
		if job.ID == id {
			q.jobs = append(q.jobs[:i], q.jobs[i+1:]...)
			return nil
		}
	}
	return nil
}
