package repository

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ridwanafazn/orchestrix/pkg/storage"
)

type StorageRepository interface {
	UploadFile(ctx context.Context, fileName string, fileContent io.Reader, contentType string) (string, error)
}

type storageRepo struct {
	s3Client *storage.S3Client
}

func NewStorageRepository(s3 *storage.S3Client) StorageRepository {
	return &storageRepo{s3Client: s3}
}

func (r *storageRepo) UploadFile(ctx context.Context, fileName string, fileContent io.Reader, contentType string) (string, error) {
	_, err := r.s3Client.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.s3Client.Bucket),
		Key:         aws.String(fileName),
		Body:        fileContent,
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload to s3: %w", err)
	}

	// Konstruksi URL (Untuk Lokal MinIO)
	// Jika di Produksi (R2), biasanya menggunakan public domain/worker URL
	fileURL := fmt.Sprintf("%s/%s/%s", r.s3Client.Endpoint, r.s3Client.Bucket, fileName)
	return fileURL, nil
}
