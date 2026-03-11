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
	return &Handler{usecases: usecases}
}

func (h *Handler) InitRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", h.HealthCheck)

	mux.HandleFunc("GET /users", h.GetsUsers)
	mux.HandleFunc("GET /users/{id}", h.GetsUserByID)
	mux.HandleFunc("POST /users", h.CreateUser)
	mux.HandleFunc("PUT /users/{id}", h.UpdateUser)
	mux.HandleFunc("DELETE /users/{id}", h.DeleteUser)

	// PATCH /users/{id}/restore  → undo a soft-delete
	mux.HandleFunc("PATCH /users/{id}/restore", h.RestoreUser)

	// GET /users/paginated?page=1&page_size=5&order_by=name&direction=asc&status=active
	mux.HandleFunc("GET /users/paginated", h.GetPaginatedUsers)

	// GET /users/cursor?limit=5&cursor=0&status=active
	mux.HandleFunc("GET /users/cursor", h.GetCursorPaginatedUsers)

	// GET /users/common-friends?user1=1&user2=2
	mux.HandleFunc("GET /users/common-friends", h.GetCommonFriends)

	// POST /users/add-friend  body: {"user_id":1,"friend_id":5}
	mux.HandleFunc("POST /users/add-friend", h.AddFriend)

	// GET /users/recommendations?user_id=1
	mux.HandleFunc("GET /users/recommendations", h.GetFriendRecommendations)

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
