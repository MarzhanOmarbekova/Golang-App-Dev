package app

import (
	"Practice3/internal/handler"
	"Practice3/internal/middleware"
	"Practice3/internal/repository"
	"Practice3/internal/repository/_postgres"
	"Practice3/internal/usecase"
	"Practice3/pkg/modules"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func Run() {
	ctx := context.Background()

	// DB config from ENV
	dbConfig := initPostgreConfig()

	_postgre := _postgres.NewPGXDialect(ctx, dbConfig)

	repositories := repository.NewRepositories(_postgre)
	usecases := usecase.NewUsecases(repositories)
	h := handler.NewHandler(usecases)

	router := h.InitRoutes()

	chainHandler := middleware.Chain(
		router,
		middleware.Logger,
		middleware.Auth,
	)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      chainHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Println("Server starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	<-quit
	log.Println("Shutting down gracefully...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Printf("Server Shutdown Failed: %+v", err)
	}

	_postgre.DB.Close()
	log.Println("Server exited properly")
}

func initPostgreConfig() *modules.PostgreConfig {
	return &modules.PostgreConfig{
		Host:        os.Getenv("DB_HOST"),
		Port:        os.Getenv("DB_PORT"),
		Username:    os.Getenv("DB_USER"),
		Password:    os.Getenv("DB_PASSWORD"),
		DBName:      os.Getenv("DB_NAME"),
		SSLMode:     "disable",
		ExecTimeout: 5 * time.Second,
	}
}
