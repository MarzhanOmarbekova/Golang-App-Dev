package handler

import (
	"Practice3/pkg/modules"
	"encoding/json"
	"net/http"
	"strconv"
)

func (h *Handler) GetsUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.usecases.GetUsers()
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (h *Handler) GetsUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid user id")
		return
	}

	user, err := h.usecases.GetUserByID(id)
	if err != nil {
		errorJSON(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req modules.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		errorJSON(w, http.StatusBadRequest, "name is required")
		return
	}

	newID, err := h.usecases.CreateUser(req)
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message": "user created successfully",
		"id":      newID,
	})
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req modules.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		errorJSON(w, http.StatusBadRequest, "name is required")
		return
	}

	if err := h.usecases.UpdateUser(id, req); err != nil {
		errorJSON(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "user updated successfully"})
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid user id")
		return
	}

	rowsAffected, err := h.usecases.DeleteUser(id)
	if err != nil {
		errorJSON(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message":       "user deleted successfully",
		"rows_affected": rowsAffected,
	})
}
