package repository

import "gorm.io/gorm"

type StreamerRepository struct {
	db *gorm.DB
}

func NewStreamerRepository(db *gorm.DB) *StreamerRepository {
	return &StreamerRepository{db: db}
}

func (r *StreamerRepository) GetDB() *gorm.DB {
	return r.db
}
