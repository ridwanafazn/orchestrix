package repository

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ridwanafazn/orchestrix/pkg/storage"
)

type StorageRepository interface {
	UploadFile(ctx context.Context, fileName string, fileContent io.Reader, contentType string) (string, error)
	GetPresignedURL(ctx context.Context, fileName string) (string, error) // Endpoint baru
}

type storageRepo struct {
	s3Client      *storage.S3Client
	presignClient *s3.PresignClient // Dibutuhkan untuk membuat Presigned URL
}

func NewStorageRepository(s3Config *storage.S3Client) StorageRepository {
	return &storageRepo{
		s3Client:      s3Config,
		presignClient: s3.NewPresignClient(s3Config.Client),
	}
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

	// Kita HANYA mengembalikan object key/path yang aman, bukan URL publik.
	return fileName, nil
}

// FUNGSI BARU: Enterprise Security Level (Temporary Access)
func (r *storageRepo) GetPresignedURL(ctx context.Context, fileName string) (string, error) {
	req, err := r.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.s3Client.Bucket),
		Key:    aws.String(fileName),
	}, s3.WithPresignExpires(15*time.Minute)) // URL kedaluwarsa dalam 15 menit

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return req.URL, nil
}
