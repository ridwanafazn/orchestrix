package repository

import (
	"context"

	"github.com/ridwanafazn/orchestrix/internal/domain"
	"gorm.io/gorm"
)

type JobRepository interface {
	Create(ctx context.Context, job *domain.Job) error
	Update(ctx context.Context, job *domain.Job) error
	GetByID(ctx context.Context, id string) (*domain.Job, error)
}

type jobRepo struct {
	db *gorm.DB
}

func NewJobRepository(db *gorm.DB) JobRepository {
	return &jobRepo{db: db}
}

func (r *jobRepo) Create(ctx context.Context, job *domain.Job) error {
	return r.db.WithContext(ctx).Create(job).Error
}

func (r *jobRepo) Update(ctx context.Context, job *domain.Job) error {
	return r.db.WithContext(ctx).Save(job).Error
}

func (r *jobRepo) GetByID(ctx context.Context, id string) (*domain.Job, error) {
	var job domain.Job
	if err := r.db.WithContext(ctx).First(&job, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &job, nil
}
