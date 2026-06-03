package domain

import "time"

const (
	JobStatusQueued     = "queued"
	JobStatusProcessing = "processing"
	JobStatusCompleted  = "completed"
	JobStatusFailed     = "failed"
)

type Job struct {
	ID          string         `json:"job_id"`
	Printer     string         `json:"printer"`
	Status      string         `json:"status"`
	Bytes       []byte         `json:"bytes,omitempty"`
	Render      map[string]any `json:"render,omitempty"`
	Error       string         `json:"error,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
}
