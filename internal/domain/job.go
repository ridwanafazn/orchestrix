package domain

import (
	"time"
)

type JobStatus string

const (
	StatusPending   JobStatus = "PENDING"
	StatusInProcess JobStatus = "PROCESSING"
	StatusCompleted JobStatus = "COMPLETED"
	StatusFailed    JobStatus = "FAILED"
)

type Job struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID    string    `gorm:"index" json:"user_id"`
	Type      string    `json:"type"`
	Status    JobStatus `gorm:"type:varchar(20);default:'PENDING'" json:"status"`
	Payload   string    `json:"payload"`
	ResultURL string    `json:"result_url"`
	ErrorLog  string    `json:"error_log"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
