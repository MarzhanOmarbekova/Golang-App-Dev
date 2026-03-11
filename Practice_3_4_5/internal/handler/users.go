package handler

import (
	"Practice3/pkg/modules"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
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
	id, err := parseID(r)
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
	writeJSON(w, http.StatusCreated, map[string]any{"message": "user created successfully", "id": newID})
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
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
	id, err := parseID(r)
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
		"message":       "user soft-deleted successfully",
		"rows_affected": rowsAffected,
	})
}

func (h *Handler) RestoreUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid user id")
		return
	}
	if err := h.usecases.RestoreUser(id); err != nil {
		errorJSON(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "user restored successfully"})
}

func (h *Handler) GetPaginatedUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page := intQueryParam(q.Get("page"), 1)
	pageSize := intQueryParam(q.Get("page_size"), 10)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	sort := modules.UserSort{
		Column:    q.Get("order_by"),
		Direction: q.Get("direction"),
	}

	filter := modules.UserFilter{
		Status: q.Get("status"),
	}

	if v := q.Get("id"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.ID = &n
		}
	}
	if v := q.Get("name"); v != "" {
		filter.Name = &v
	}
	if v := q.Get("email"); v != "" {
		filter.Email = &v
	}
	if v := q.Get("gender"); v != "" {
		filter.Gender = &v
	}
	if v := q.Get("birth_date"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			filter.BirthDate = &t
		}
	}

	result, err := h.usecases.GetPaginatedUsers(page, pageSize, filter, sort)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) GetCursorPaginatedUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	limit := intQueryParam(q.Get("limit"), 10)
	cursor := intQueryParam(q.Get("cursor"), 0)

	filter := modules.UserFilter{
		Status: q.Get("status"),
	}
	if v := q.Get("name"); v != "" {
		filter.Name = &v
	}
	if v := q.Get("gender"); v != "" {
		filter.Gender = &v
	}

	result, err := h.usecases.GetCursorPaginatedUsers(cursor, limit, filter)
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) GetCommonFriends(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	user1Str := q.Get("user1")
	user2Str := q.Get("user2")
	if user1Str == "" || user2Str == "" {
		errorJSON(w, http.StatusBadRequest, "user1 and user2 query params are required")
		return
	}

	user1, err := strconv.Atoi(user1Str)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid user1 id")
		return
	}
	user2, err := strconv.Atoi(user2Str)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid user2 id")
		return
	}
	if user1 == user2 {
		errorJSON(w, http.StatusBadRequest, "user1 and user2 must be different")
		return
	}

	friends, err := h.usecases.GetCommonFriends(user1, user2)
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user1_id":       user1,
		"user2_id":       user2,
		"common_friends": friends,
		"count":          len(friends),
	})
}

func (h *Handler) AddFriend(w http.ResponseWriter, r *http.Request) {
	var req modules.AddFriendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.UserID == 0 || req.FriendID == 0 {
		errorJSON(w, http.StatusBadRequest, "user_id and friend_id are required")
		return
	}
	if req.UserID == req.FriendID {
		errorJSON(w, http.StatusBadRequest, "user_id and friend_id must be different")
		return
	}

	if err := h.usecases.AddFriend(req.UserID, req.FriendID); err != nil {
		errorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"message":   "friendship added successfully (both directions)",
		"user_id":   req.UserID,
		"friend_id": req.FriendID,
	})
}

func (h *Handler) GetFriendRecommendations(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		errorJSON(w, http.StatusBadRequest, "user_id query param is required")
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	recs, err := h.usecases.GetFriendRecommendations(userID)
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user_id":         userID,
		"recommendations": recs,
		"count":           len(recs),
	})
}

func parseID(r *http.Request) (int, error) {
	return strconv.Atoi(r.PathValue("id"))
}

func intQueryParam(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return n
}
