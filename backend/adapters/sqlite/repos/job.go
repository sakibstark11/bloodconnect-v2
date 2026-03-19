package repos

import (
	"context"

	"bloodconnect/adapters/sqlite/models"
	"bloodconnect/application"
	"bloodconnect/application/domain"

	"gorm.io/gorm"
)

type jobQueue struct {
	db *gorm.DB
}

func NewJobQueue(db *gorm.DB) application.JobQueue {
	return &jobQueue{db: db}
}

func (q *jobQueue) Enqueue(ctx context.Context, job *domain.Job) error {
	return q.db.WithContext(ctx).Create(models.JobFromDomain(job)).Error
}

func (q *jobQueue) FetchNextAvailable(ctx context.Context) (*domain.Job, error) {
	now := domain.Now()
	var m models.Job
	res := q.db.WithContext(ctx).
		Where("status = ? AND run_at <= ?", string(domain.JobStatusPending), string(now)).
		Order("created_at asc").
		First(&m)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, res.Error
	}
	return m.ToDomain(), nil
}

func (q *jobQueue) MarkStatus(ctx context.Context, id string, status domain.JobStatus) error {
	return q.db.WithContext(ctx).Model(&models.Job{}).Where("id = ?", id).Update("status", string(status)).Error
}

func (q *jobQueue) Delete(ctx context.Context, id string) error {
	return q.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Job{}).Error
}
