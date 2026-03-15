package domain

import "time"

// JobStatus track the background processor status
type JobStatus string

const (
	JobStatusPending    JobStatus = "Pending"
	JobStatusProcessing JobStatus = "Processing"
	JobStatusCompleted  JobStatus = "Completed"
	JobStatusFailed     JobStatus = "Failed"
	JobStatusDelayed    JobStatus = "Delayed" // For jobs needed far in the future
)

// JobType represents what kind of processing this job does
type JobType string

const (
	JobTypeWaveSearch JobType = "wave_search"
)

// Job represents a background task to process, specifically donor wave searches
type Job struct {
	ID        string
	Type      JobType
	Payload   string // JSON structure containing RequestID, CurrentRing (K), etc.
	Status    JobStatus
	RunAt     time.Time // When it's allowed to run next
	CreatedAt time.Time
	UpdatedAt time.Time
}

// WaveSearchPayload represents the JSON payload stored in Job.Payload
type WaveSearchPayload struct {
	RequestID         string `json:"request_id"`
	CurrentRing       int    `json:"current_ring"`
	CenterHex         string `json:"center_hex"`
	UsersLeftToSearch int    `json:"users_left_to_search"`
}
