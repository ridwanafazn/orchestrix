package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ridwanafazn/orchestrix/internal/repository"
	"github.com/ridwanafazn/orchestrix/pkg/rabbitmq"
)

type WorkerUsecase interface {
	ProcessJob(ctx context.Context, payload []byte) error
}

type workerUsecase struct {
	jobRepo   repository.JobRepository
	rmqClient *rabbitmq.RabbitMQClient
	workerID  string
}

func NewWorkerUsecase(jr repository.JobRepository, rmq *rabbitmq.RabbitMQClient) WorkerUsecase {
	return &workerUsecase{
		jobRepo:   jr,
		rmqClient: rmq,
		workerID:  fmt.Sprintf("worker-node-%d", time.Now().UnixNano()%1000),
	}
}

func (u *workerUsecase) ProcessJob(ctx context.Context, payload []byte) error {
	var event struct {
		JobID        string `json:"job_id"`
		ChaosInject  bool   `json:"chaos_inject"`
		HeavyCompute bool   `json:"heavy_compute"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	job, err := u.jobRepo.GetByID(ctx, event.JobID)
	if err != nil {
		return fmt.Errorf("job %s not found: %w", event.JobID, err)
	}

	if string(job.Status) == "COMPLETED" || string(job.Status) == "FAILED" {
		log.Printf("⏩ Job %s is already %s, skipping.", job.ID, job.Status)
		return nil
	}

	// 1. UPDATE STATUS -> PROCESSING
	now := time.Now()
	job.Status = "PROCESSING"
	job.WorkerID = u.workerID
	job.Progress = 10
	job.StartedAt = &now
	_ = u.jobRepo.Update(ctx, job)

	log.Printf("⏳ [PROCESSING] %s menarik Job %s...", u.workerID, job.ID)
	u.publishNotification(job.UserID, job.ID, "PROCESSING", 10, "")

	// 2. SIMULASI PEKERJAAN BERAT BERTAHAP
	delay := 800 * time.Millisecond
	if event.HeavyCompute {
		log.Printf("🧱 [HEAVY COMPUTE] %s executing matrix calculation for Job %s...", u.workerID, job.ID)
		delay = 4 * time.Second // 4 detik * 3 tahap = 12 detik penahanan
	}

	steps := []int{35, 65, 90}
	for _, progress := range steps {
		time.Sleep(delay)
		job.Progress = progress
		_ = u.jobRepo.Update(ctx, job)
		u.publishNotification(job.UserID, job.ID, "PROCESSING", progress, "")
	}

	// 3. CHAOS ENGINEERING TRIGGER
	if event.ChaosInject {
		time.Sleep(500 * time.Millisecond)
		job.Status = "FAILED"
		job.ErrorLog = "Simulated Fatal Corrupt Data. Worker Panicked."
		job.Progress = 0
		_ = u.jobRepo.Update(ctx, job)

		log.Printf("💥 [FAILED] Chaos injected on Job %s!", job.ID)
		u.publishNotification(job.UserID, job.ID, "FAILED", 0, "")

		// Melempar error absolut agar RabbitMQ men-Nack dan membuang ke DLX
		return fmt.Errorf("chaos engineered failure on worker %s", u.workerID)
	}

	// 4. UPDATE STATUS -> COMPLETED
	completedTime := time.Now()
	job.Status = "COMPLETED"
	job.Progress = 100
	job.CompletedAt = &completedTime
	job.ResultURL = "minio-vault:/" + job.ID + ".png"

	if err := u.jobRepo.Update(ctx, job); err != nil {
		return fmt.Errorf("failed to set job to completed: %w", err)
	}

	log.Printf("✅ [COMPLETED] Job %s berhasil diproses oleh %s!", job.ID, u.workerID)
	u.publishNotification(job.UserID, job.ID, "COMPLETED", 100, job.ResultURL)

	return nil
}

func (u *workerUsecase) publishNotification(userID, jobID, status string, progress int, resultURL string) {
	notifPayload := map[string]interface{}{
		"user_id":    userID,
		"job_id":     jobID,
		"status":     status,
		"progress":   progress,
		"worker_id":  u.workerID,
		"result_url": resultURL,
	}
	notifBytes, _ := json.Marshal(notifPayload)

	if err := u.rmqClient.Publish("notification_queue", notifBytes); err != nil {
		log.Printf("⚠️ Failed to publish notification for job %s: %v", jobID, err)
	}
}
