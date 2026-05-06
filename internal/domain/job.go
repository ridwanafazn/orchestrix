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

	// --- ENRICHED TELEMETRY FIELDS FOR 3D OBSERVABILITY ---
	WorkerID    string     `json:"worker_id"`    // Identifikasi Node Worker mana yang mengambil job ini
	Progress    int        `json:"progress"`     // Untuk update % di UI Logger
	StartedAt   *time.Time `json:"started_at"`   // Waktu presisi saat ditarik dari antrean
	CompletedAt *time.Time `json:"completed_at"` // Waktu presisi saat selesai
	// ------------------------------------------------------

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
