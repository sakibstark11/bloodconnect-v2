package sqlite

import (
	"context"

	"gorm.io/gorm"

	"github.com/sakibalam/bloodconnect/domain"
	"github.com/sakibalam/bloodconnect/ports"
)

type jobQueue struct {
	db *gorm.DB
}

func NewJobQueue(db *gorm.DB) ports.JobQueue {
	return &jobQueue{
		db: db,
	}
}

func (q *jobQueue) Enqueue(ctx context.Context, job *domain.Job) error {
	m := fromDomainJob(job)
	res := q.db.WithContext(ctx).Create(m)
	return res.Error
}

func (q *jobQueue) FetchNextAvailable(ctx context.Context) (*domain.Job, error) {
	var m jobModel
	// Find the oldest pending job that is ready to run
	// In SQLite, we can lock by using a transaction or just simple status updates.
	// We'll do a simple find for now, worker pool will need to handle concurrency 
	// (e.g., Update status to Processing where id=X and status=Pending)
	
	res := q.db.WithContext(ctx).Where("status = ? AND run_at <= CURRENT_TIMESTAMP", domain.JobStatusPending).Order("created_at asc").First(&m)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, nil // No jobs available
		}
		return nil, res.Error
	}
	return m.toDomain(), nil
}

func (q *jobQueue) MarkStatus(ctx context.Context, id string, status domain.JobStatus) error {
	res := q.db.WithContext(ctx).Model(&jobModel{}).Where("id = ?", id).Update("status", string(status))
	return res.Error
}

func (q *jobQueue) Delete(ctx context.Context, id string) error {
	res := q.db.WithContext(ctx).Where("id = ?", id).Delete(&jobModel{})
	return res.Error
}
