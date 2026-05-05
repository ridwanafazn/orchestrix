package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ridwanafazn/orchestrix/internal/domain"
	"github.com/ridwanafazn/orchestrix/internal/repository"
	"github.com/ridwanafazn/orchestrix/pkg/rabbitmq"
)

type WorkerUsecase interface {
	ProcessJob(ctx context.Context, payload []byte) error
}

type workerUsecase struct {
	jobRepo   repository.JobRepository
	rmqClient *rabbitmq.RabbitMQClient
}

// UPDATE: Tambahkan injeksi RabbitMQ ke constructor
func NewWorkerUsecase(jr repository.JobRepository, rmq *rabbitmq.RabbitMQClient) WorkerUsecase {
	return &workerUsecase{jobRepo: jr, rmqClient: rmq}
}

func (u *workerUsecase) ProcessJob(ctx context.Context, payload []byte) error {
	var event struct {
		JobID string `json:"job_id"`
		Type  string `json:"type"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	job, err := u.jobRepo.GetByID(ctx, event.JobID)
	if err != nil {
		return fmt.Errorf("job %s not found: %w", event.JobID, err)
	}

	if job.Status == domain.StatusCompleted || job.Status == domain.StatusFailed {
		log.Printf("⏩ Job %s is already %s, skipping.", job.ID, job.Status)
		return nil
	}

	job.Status = domain.StatusInProcess
	_ = u.jobRepo.Update(ctx, job)
	log.Printf("⏳ [PROCESSING] Job %s is starting...", job.ID)

	time.Sleep(5 * time.Second)

	job.Status = domain.StatusCompleted
	if err := u.jobRepo.Update(ctx, job); err != nil {
		return fmt.Errorf("failed to set job to completed: %w", err)
	}

	log.Printf("✅ [COMPLETED] Job %s processed successfully!", job.ID)

	// ==========================================
	// 6. EVENT PUBLISHER: Kirim notifikasi ke API Gateway
	// ==========================================
	notifPayload := map[string]string{
		"job_id": job.ID,
		"status": string(job.Status),
	}
	notifBytes, _ := json.Marshal(notifPayload)

	if err := u.rmqClient.Publish("notification_queue", notifBytes); err != nil {
		log.Printf("⚠️ Failed to publish notification for job %s: %v", job.ID, err)
	}

	return nil
}
