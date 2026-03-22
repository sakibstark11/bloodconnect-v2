package models

import (
	"encoding/json"

	"bloodconnect/application/domain"

	"gorm.io/datatypes"
)

type Job struct {
	ID        string            `gorm:"primaryKey;size:64"`
	Type      string            `gorm:"index;size:50"`
	Payload   datatypes.JSONMap `gorm:"type:jsonb"`
	Status    string            `gorm:"index;size:20"`
	RunAt     string            `gorm:"size:25"`
	CreatedAt string            `gorm:"size:25"`
	UpdatedAt string            `gorm:"size:25"`
}

func (m *Job) ToDomain() *domain.Job {
	payload := ""
	if len(m.Payload) > 0 {
		bytes, _ := json.Marshal(m.Payload)
		payload = string(bytes)
	}
	return &domain.Job{
		ID:        domain.JobID(m.ID),
		Type:      domain.JobType(m.Type),
		Payload:   payload,
		Status:    domain.JobStatus(m.Status),
		RunAt:     domain.ISOTimestamp(m.RunAt),
		CreatedAt: domain.ISOTimestamp(m.CreatedAt),
		UpdatedAt: domain.ISOTimestamp(m.UpdatedAt),
	}
}

func JobFromDomain(j *domain.Job) *Job {
	// Parse the JSON string payload back into a JSONMap for native JSONB storage.
	var jmap datatypes.JSONMap
	if j.Payload != "" {
		_ = jmap.UnmarshalJSON([]byte(j.Payload))
	}
	return &Job{
		ID:        string(j.ID),
		Type:      string(j.Type),
		Payload:   jmap,
		Status:    string(j.Status),
		RunAt:     string(j.RunAt),
		CreatedAt: string(j.CreatedAt),
		UpdatedAt: string(j.UpdatedAt),
	}
}
