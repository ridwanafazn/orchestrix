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
		log.Println("⚠️  No .env file found")
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

	// Channel sebagai penahan agar CLI tidak exit
	forever := make(chan struct{})

	log.Println("👷 Worker Node Online! Menunggu antrean pekerjaan...")

	// 4. Goroutine Pool Listener
	go func() {
		for msg := range msgs {
			// Spawn goroutine baru untuk setiap pesan agar berjalan paralel
			// secara non-blocking, namun tetap dibatasi oleh Prefetch=10.
			go func(delivery amqp.Delivery) {
				ctx := context.Background()

				err := workerUsecase.ProcessJob(ctx, delivery.Body)
				if err != nil {
					log.Printf("❌ Error processing job: %v", err)
					// Reject (Nack). Argumen (multiple: false, requeue: false).
					// Karena requeue=false, pesan ini akan dibuang (atau masuk Dead Letter Exchange jika disetup)
					delivery.Nack(false, false)
				} else {
					// Manual Ack: Mengabari RabbitMQ bahwa tugas aman diselesaikan,
					// hapus permanen dari antrean.
					delivery.Ack(false)
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
