package models

import (
	"time"

	"bloodconnect/application/domain"
)

// Job is the GORM model for domain.Job
type Job struct {
	ID        string    `gorm:"primaryKey;size:64"`
	Type      string    `gorm:"index;size:50"`
	Payload   string    `gorm:"type:text"`
	Status    string    `gorm:"index;size:20"`
	RunAt     time.Time `gorm:"index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *Job) ToDomain() *domain.Job {
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

func JobFromDomain(j *domain.Job) *Job {
	return &Job{
		ID:        j.ID,
		Type:      string(j.Type),
		Payload:   j.Payload,
		Status:    string(j.Status),
		RunAt:     j.RunAt,
		CreatedAt: j.CreatedAt,
		UpdatedAt: j.UpdatedAt,
	}
}
