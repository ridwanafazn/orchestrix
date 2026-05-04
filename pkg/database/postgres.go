package database

import (
	"log"

	"github.com/ridwanafazn/orchestrix/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitPostgres(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = db.AutoMigrate(&domain.Job{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}
