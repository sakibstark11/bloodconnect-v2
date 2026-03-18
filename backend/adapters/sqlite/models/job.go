package models

import (
	"bloodconnect/application/domain"
)

// Job is the GORM model for domain.Job
type Job struct {
	ID        string `gorm:"primaryKey;size:64"`
	Type      string `gorm:"index;size:50"`
	Payload   string `gorm:"type:text"`
	Status    string `gorm:"index;size:20"`
	RunAt     string `gorm:"size:25"`
	CreatedAt string `gorm:"size:25"`
	UpdatedAt string `gorm:"size:25"`
}

func (m *Job) ToDomain() *domain.Job {
	return &domain.Job{
		ID:        m.ID,
		Type:      domain.JobType(m.Type),
		Payload:   m.Payload,
		Status:    domain.JobStatus(m.Status),
		RunAt:     domain.ISOTimestamp(m.RunAt),
		CreatedAt: domain.ISOTimestamp(m.CreatedAt),
		UpdatedAt: domain.ISOTimestamp(m.UpdatedAt),
	}
}

func JobFromDomain(j *domain.Job) *Job {
	return &Job{
		ID:        j.ID,
		Type:      string(j.Type),
		Payload:   j.Payload,
		Status:    string(j.Status),
		RunAt:     string(j.RunAt),
		CreatedAt: string(j.CreatedAt),
		UpdatedAt: string(j.UpdatedAt),
	}
}
