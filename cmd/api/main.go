package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/ridwanafazn/orchestrix/internal/repository"
	"github.com/ridwanafazn/orchestrix/pkg/database"
	"github.com/ridwanafazn/orchestrix/pkg/storage"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// 1. Init Infra
	_ = database.InitPostgres(os.Getenv("DB_URL"))

	s3Infra := storage.InitS3(
		os.Getenv("S3_ENDPOINT"),
		os.Getenv("S3_REGION"),
		os.Getenv("S3_BUCKET"),
		os.Getenv("S3_ACCESS_KEY"),
		os.Getenv("S3_SECRET_KEY"),
	)

	// 2. Init Repository
	storageRepo := repository.NewStorageRepository(s3Infra)

	// 3. Test Upload (Smoke Test)
	testContent := "Halo Wang, ini file testing dari Orchestrix!"
	reader := strings.NewReader(testContent)

	url, err := storageRepo.UploadFile(context.Background(), "test-connection.txt", reader, "text/plain")
	if err != nil {
		log.Fatalf("❌ Upload Gagal: %v", err)
	}

	log.Printf("✅ Smoke Test Berhasil! File diupload ke: %s", url)
}
