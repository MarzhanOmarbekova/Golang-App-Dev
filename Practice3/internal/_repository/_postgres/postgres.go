package _postgres

import (
	"Practice3/pkg/modules"
	"context"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	//"github.com/golang-migrate/migrate/v4/source"
	"github.com/jmoiron/sqlx"
	//"github.com/lib/pq"
	//"github.com/golang-migrate/migrate/v4/database/_postgres"
	//"github.com/golang-migrate/migrate/v4/source/file"
)

type Dialect struct {
	DB *sqlx.DB
}

func NewPGXDialect(ctx context.Context, cfg *modules.PostgreConfig) *Dialect {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := sqlx.Connect("posgres", dsn)

	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	AutoMigrate(cfg)

	return &Dialect{
		DB: db,
	}
}

func AutoMigrate(cfg *modules.PostgreConfig) {
	sourceURL := "file://database/migrations"
	databaseURL := fmt.Sprintf("_postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)

	m, err := migrate.New(sourceURL, databaseURL)

	if err != nil {
		panic(err)
	}

	err = m.Up()

	if err != nil && err != migrate.ErrNoChange {
		panic(err)
	}
}
