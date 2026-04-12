package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"practice-7/config"
	v1 "practice-7/internal/controller/http/v1"
	"practice-7/internal/usecase"
	"practice-7/internal/usecase/repo"
	"practice-7/pkg/logger"
	"practice-7/pkg/postgres"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func Run(cfg *config.Config) {
	l := logger.New("info")

	// Postgres
	pg, err := postgres.New(cfg.PgURL)
	if err != nil {
		log.Fatalf("postgres connect: %v", err)
	}
	l.Info("Postgres connected")

	// Redis
	var rdb *redis.Client
	if cfg.RedisAddr != "" {
		rdb = redis.NewClient(&redis.Options{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		})
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := rdb.Ping(ctx).Err(); err != nil {
			l.Error("Redis not available: %v - running WITHOUT Redis (stateless mode)", err)
			rdb = nil
		} else {
			l.Info("Redis connected")
		}
	}

	// Use case
	userRepo := repo.NewUserRepo(pg)
	userUseCase := usecase.NewUserUseCase(userRepo, rdb)

	// HTTP
	handler := gin.Default()
	v1.NewRouter(handler, userUseCase, l, rdb)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: handler,
	}

	// Start server
	go func() {
		l.Info("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	l.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	l.Info("Server exited")
}
