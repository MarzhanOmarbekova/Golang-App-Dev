package main

import (
	"log"
	"net/http"
	"task_api/internal/handlers"
	"task_api/internal/middleware"
	"task_api/internal/storage"
)

func main() {
	store := storage.NewTaskStorage()
	handler := handlers.TaskHandler{Store: store}

	mux := http.NewServeMux()
	mux.Handle("/tasks", middleware.APIKey(middleware.Logging(http.HandlerFunc(handler.Tasks))))

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
