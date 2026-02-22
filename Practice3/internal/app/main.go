package app

import (
	"Practice3/internal/handler"
	"Practice3/internal/middleware"
	"Practice3/internal/repository"
	"Practice3/internal/repository/_postgres"
	"Practice3/internal/usecase"
	"Practice3/pkg/modules"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

func Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbConfig := initPostgreConfig()
	_postgre := _postgres.NewPGXDialect(ctx, dbConfig)

	fmt.Println("Connected to database: ", _postgre)

	repositories := repository.NewRepositories(_postgre)
	usecases := usecase.NewUsecases(repositories)
	h := handler.NewHandler(usecases)

	router := h.InitRoutes()

	chainHandler := middleware.Chain(router, middleware.Logger, middleware.Auth)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      chainHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Server starting on :8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func initPostgreConfig() *modules.PostgreConfig {
	return &modules.PostgreConfig{
		Host:        "localhost",
		Port:        "5432",
		Username:    "postgres",
		Password:    "marzhan06",
		DBName:      "mydb",
		SSLMode:     "disable",
		ExecTimeout: 5 * time.Second,
	}
}
