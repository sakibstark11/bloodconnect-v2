package domain

type JobID string

type JobStatus string

const (
	JobStatusPending    JobStatus = "Pending"
	JobStatusProcessing JobStatus = "Processing"
	JobStatusCompleted  JobStatus = "Completed"
	JobStatusFailed     JobStatus = "Failed"
	JobStatusDelayed    JobStatus = "Delayed"
)

type JobType string

const (
	JobTypeWaveSearch     JobType = "wave_search"
	JobTypeCheckResponses JobType = "check_responses"
	JobTypeNotification   JobType = "notification"
)

type Job struct {
	ID        JobID
	Type      JobType
	Payload   interface{}
	Status    JobStatus
	RunAt     ISOTimestamp
	CreatedAt ISOTimestamp
	UpdatedAt ISOTimestamp
}

type WaveSearchPayload struct {
	RequestID   RequestID `json:"request_id"`
	CurrentRing int       `json:"current_ring"`
	RetryCount  int       `json:"retry_count"`
}
