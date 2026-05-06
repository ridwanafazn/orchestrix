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
	uniqueFileName := fmt.Sprintf("%d_%s", time.Now().Unix(), fileName)
	fileURL, err := u.storageRepo.UploadFile(ctx, uniqueFileName, file, contentType)
	if err != nil {
		job.Status = domain.StatusFailed
		job.ErrorLog = err.Error()
		_ = u.jobRepo.Update(ctx, job)
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	job.ResultURL = fileURL
	if err := u.jobRepo.Update(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to update job with url: %w", err)
	}

	// Tangkap flag dari Context
	chaosInject := false
	if val := ctx.Value("chaos_inject"); val != nil {
		chaosInject = val.(bool)
	}

	heavyCompute := false
	if val := ctx.Value("heavy_compute"); val != nil {
		heavyCompute = val.(bool)
	}

	// 3. Publish Event ke RabbitMQ
	eventPayload := map[string]interface{}{
		"job_id":        job.ID,
		"type":          job.Type,
		"chaos_inject":  chaosInject,
		"heavy_compute": heavyCompute,
	}
	payloadBytes, _ := json.Marshal(eventPayload)

	if err := u.rmqClient.Publish("kyc_queue", payloadBytes); err != nil {
		job.ErrorLog = "Failed to publish to message broker"
		_ = u.jobRepo.Update(ctx, job)
	}

	return job, nil
}
