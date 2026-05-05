package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/ridwanafazn/orchestrix/internal/domain"
	"github.com/ridwanafazn/orchestrix/internal/repository"
	"github.com/ridwanafazn/orchestrix/pkg/rabbitmq"
)

type JobUsecase interface {
	ProcessKYCUpload(ctx context.Context, userID string, fileName string, file io.Reader, contentType string) (*domain.Job, error)
}

type jobUsecase struct {
	jobRepo     repository.JobRepository
	storageRepo repository.StorageRepository
	rmqClient   *rabbitmq.RabbitMQClient
}

func NewJobUsecase(jr repository.JobRepository, sr repository.StorageRepository, rmq *rabbitmq.RabbitMQClient) JobUsecase {
	return &jobUsecase{
		jobRepo:     jr,
		storageRepo: sr,
		rmqClient:   rmq,
	}
}

func (u *jobUsecase) ProcessKYCUpload(ctx context.Context, userID string, fileName string, file io.Reader, contentType string) (*domain.Job, error) {
	// 1. Catat Job ke Database dengan status PENDING
	job := &domain.Job{
		UserID: userID,
		Type:   "KYC_COMPRESSION",
		Status: domain.StatusPending,
	}

	if err := u.jobRepo.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create job record: %w", err)
	}

	// 2. Upload file fisik ke MinIO/R2
	// Tambahkan prefix timestamp agar nama file unik
	uniqueFileName := fmt.Sprintf("%d_%s", time.Now().Unix(), fileName)
	fileURL, err := u.storageRepo.UploadFile(ctx, uniqueFileName, file, contentType)
	if err != nil {
		// Jika upload gagal, set status ke FAILED
		job.Status = domain.StatusFailed
		job.ErrorLog = err.Error()
		_ = u.jobRepo.Update(ctx, job)
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Update DB dengan URL file
	job.ResultURL = fileURL
	if err := u.jobRepo.Update(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to update job with url: %w", err)
	}

	// 3. Publish Event ke RabbitMQ (Worker akan memprosesnya nanti)
	eventPayload := map[string]string{
		"job_id": job.ID,
		"type":   job.Type,
	}
	payloadBytes, _ := json.Marshal(eventPayload)

	if err := u.rmqClient.Publish("kyc_queue", payloadBytes); err != nil {
		// Toleransi kegagalan: Catat error tapi jangan gagalkan request ke user
		job.ErrorLog = "Failed to publish to message broker"
		_ = u.jobRepo.Update(ctx, job)
	}

	return job, nil
}
