package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/ridwanafazn/orchestrix/internal/repository"
	"github.com/ridwanafazn/orchestrix/internal/usecase"
	"github.com/ridwanafazn/orchestrix/pkg/database"
	"github.com/ridwanafazn/orchestrix/pkg/rabbitmq"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using system environment variables")
	}

	// 1. Init Infra
	db := database.InitPostgres(os.Getenv("DB_URL"))
	rmqInfra := rabbitmq.InitRabbitMQ(os.Getenv("RABBITMQ_URL"))
	defer rmqInfra.Close()

	// 2. Init Repo & Usecase
	jobRepo := repository.NewJobRepository(db)
	workerUsecase := usecase.NewWorkerUsecase(jobRepo, rmqInfra)

	// 3. Mulai Konsumsi Pesan (Prefetch: 10 worker concurrency limit)
	msgs, err := rmqInfra.Consume("kyc_queue", 10)
	if err != nil {
		log.Fatalf("❌ Gagal listen ke RabbitMQ: %v", err)
	}

	forever := make(chan struct{})

	log.Println("👷 Worker Node Online! Menunggu antrean pekerjaan dari kyc_queue...")

	// 4. Goroutine Pool Listener
	go func() {
		for msg := range msgs {
			// Spawn goroutine baru untuk setiap pesan agar berjalan paralel
			go func(delivery amqp.Delivery) {
				ctx := context.Background()

				err := workerUsecase.ProcessJob(ctx, delivery.Body)
				if err != nil {
					log.Printf("💥 Error processing job (Routing to DLX): %v", err)
					// WAJIB requeue=false agar RabbitMQ melemparnya ke Dead Letter Exchange
					delivery.Nack(false, false)
				} else {
					delivery.Ack(false) // Sukses, hapus permanen dari antrean
				}
			}(msg)
		}
	}()

	// 5. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🔄 Worker node is shutting down...")
	close(forever)
}
