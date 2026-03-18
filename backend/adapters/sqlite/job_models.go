package sqlite

import (
	"time"

	"github.com/sakibalam/bloodconnect/domain"
)

type jobModel struct {
	ID        string `gorm:"primaryKey;size:64"`
	Type      string `gorm:"index;size:50"`
	Payload   string `gorm:"type:text"`
	Status    string `gorm:"index;size:20"`
	RunAt     time.Time `gorm:"index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *jobModel) toDomain() *domain.Job {
	return &domain.Job{
		ID:        m.ID,
		Type:      domain.JobType(m.Type),
		Payload:   m.Payload,
		Status:    domain.JobStatus(m.Status),
		RunAt:     m.RunAt,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func fromDomainJob(j *domain.Job) *jobModel {
	return &jobModel{
		ID:        j.ID,
		Type:      string(j.Type),
		Payload:   j.Payload,
		Status:    string(j.Status),
		RunAt:     j.RunAt,
		CreatedAt: j.CreatedAt,
		UpdatedAt: j.UpdatedAt,
	}
}
