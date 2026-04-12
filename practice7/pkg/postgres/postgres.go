package postgres

import (
	"fmt"
	"practice-7/internal/entity"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Postgres struct {
	Conn *gorm.DB
}

func New(url string) (*Postgres, error) {
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("postgres connect: %w", err)
	}

	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		return nil, fmt.Errorf("uuid extension: %w", err)
	}

	if err := db.AutoMigrate(&entity.User{}); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	return &Postgres{Conn: db}, nil
}
