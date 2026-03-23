package memory

import (
	"context"
	"sync"
	"time"

	"bloodconnect/application"
	"bloodconnect/application/domain"
)

type inMemoryJobQueue struct {
	mu     sync.Mutex
	jobs   map[domain.JobType][]*domain.Job
	chans  map[domain.JobType]chan *domain.Job
}

func NewJobQueue() application.JobQueue {
	return &inMemoryJobQueue{
		jobs:  make(map[domain.JobType][]*domain.Job),
		chans: make(map[domain.JobType]chan *domain.Job),
	}
}

func (q *inMemoryJobQueue) Enqueue(ctx context.Context, job *domain.Job) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.jobs[job.Type] = append(q.jobs[job.Type], job)

	delay := time.Until(job.RunAt.ToTime())
	if delay <= 0 {
		q.deliver(job)
	} else {
		time.AfterFunc(delay, func() {
			q.mu.Lock()
			defer q.mu.Unlock()
			q.deliver(job)
		})
	}

	return nil
}

func (q *inMemoryJobQueue) deliver(job *domain.Job) {
	if c, ok := q.chans[job.Type]; ok {
		select {
		case c <- job:
		default:
			// Buffer full
		}
	}
}

func (q *inMemoryJobQueue) Consume(ctx context.Context, jobType domain.JobType) (<-chan *domain.Job, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, ok := q.chans[jobType]; !ok {
		ch := make(chan *domain.Job, 100)
		q.chans[jobType] = ch
		
		// Drain existing jobs into the channel
	drainLoop:
		for _, job := range q.jobs[jobType] {
			select {
			case ch <- job:
			default:
				// Buffer full, stop draining
				break drainLoop
			}
		}
		// Clear jobs as they are now in the channel
		q.jobs[jobType] = nil
	}
	
	return q.chans[jobType], nil
}

func (q *inMemoryJobQueue) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, ch := range q.chans {
		close(ch)
	}
	return nil
}
