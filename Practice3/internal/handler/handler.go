package handler

import (
	"Practice3/internal/usecase"
	"encoding/json"
	"net/http"
)

type Handler struct {
	usecases *usecase.Usecases
}

func NewHandler(usecases *usecase.Usecases) *Handler {
	return &Handler{
		usecases: usecases,
	}
}

func (h *Handler) InitRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", h.HealthCheck)

	mux.HandleFunc("GET /users", h.GetsUsers)
	mux.HandleFunc("GET /users/{id}", h.GetsUserByID)
	mux.HandleFunc("POST /users", h.CreateUser)
	mux.HandleFunc("PUT /users/{id}", h.UpdateUser)
	mux.HandleFunc("DELETE /users/{id}", h.DeleteUser)

	return mux
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func errorJSON(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
