package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	deliveryHttp "github.com/ridwanafazn/orchestrix/internal/delivery/http"
	ws "github.com/ridwanafazn/orchestrix/internal/delivery/websocket"
	"github.com/ridwanafazn/orchestrix/internal/repository"
	"github.com/ridwanafazn/orchestrix/internal/usecase"
	"github.com/ridwanafazn/orchestrix/pkg/database"
	"github.com/ridwanafazn/orchestrix/pkg/rabbitmq"
	"github.com/ridwanafazn/orchestrix/pkg/storage"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, relying on system variables")
	}

	db := database.InitPostgres(os.Getenv("DB_URL"))

	s3Infra := storage.InitS3(
		os.Getenv("S3_ENDPOINT"),
		os.Getenv("S3_REGION"),
		os.Getenv("S3_BUCKET"),
		os.Getenv("S3_ACCESS_KEY"),
		os.Getenv("S3_SECRET_KEY"),
	)

	rmqInfra := rabbitmq.InitRabbitMQ(os.Getenv("RABBITMQ_URL"))
	defer rmqInfra.Close()

	// ==========================================
	// 2. Inisialisasi WebSocket Hub
	// ==========================================
	hub := ws.NewHub()
	go hub.Run()

	jobRepo := repository.NewJobRepository(db)
	storageRepo := repository.NewStorageRepository(s3Infra)

	jobUsecase := usecase.NewJobUsecase(jobRepo, storageRepo, rmqInfra)
	jobHandler := deliveryHttp.NewJobHandler(jobUsecase)

	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	deliveryHttp.InitRouter(engine, jobHandler, hub) // Inject hub ke Router

	// ==========================================
	// 3. Listener RabbitMQ -> WebSocket Broadcaster
	// ==========================================
	go func() {
		msgs, err := rmqInfra.Consume("notification_queue", 10)
		if err != nil {
			log.Printf("⚠️ Gagal listen ke notification_queue: %v", err)
			return
		}

		log.Println("📡 Notification Gateway Listening...")
		for msg := range msgs {
			// Lempar payload RabbitMQ langsung ke semua klien WS
			hub.Broadcast <- msg.Body
			msg.Ack(false)
		}
	}()

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: engine,
	}

	go func() {
		log.Printf("🚀 Server is running on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🔄 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("❌ Server forced to shutdown:", err)
	}

	log.Println("✅ Server exiting cleanly")
}
