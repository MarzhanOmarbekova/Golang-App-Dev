package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"task_api/internal/storage"
)

type TaskHandler struct {
	Store *storage.TaskStorage
}

func (h *TaskHandler) Tasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {

	case http.MethodGet:
		idStr := r.URL.Query().Get("id")
		doneStr := r.URL.Query().Get("done")

		if idStr != "" {
			id, err := strconv.Atoi(idStr)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid id"})
				return
			}
			task, ok := h.Store.GetByID(id)
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "task not found"})
				return
			}
			json.NewEncoder(w).Encode(task)
			return
		}

		var doneFilter *bool
		if doneStr != "" {
			val, err := strconv.ParseBool(doneStr)
			if err == nil {
				doneFilter = &val
			}
		}
		json.NewEncoder(w).Encode(h.Store.GetAll(doneFilter))

	case http.MethodPost:
		var body struct {
			Title string `json:"title"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Title == "" || len(body.Title) > 100 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid title"})
			return
		}
		task := h.Store.Create(body.Title)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)

	case http.MethodPatch:
		idStr := r.URL.Query().Get("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid id"})
			return
		}

		var body struct {
			Done bool `json:"done"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if !h.Store.UpdateDone(id, body.Done) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "task not found"})
			return
		}
		json.NewEncoder(w).Encode(map[string]bool{"updated": true})

	case http.MethodDelete:
		idStr := r.URL.Query().Get("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid id"})
			return
		}
		if !h.Store.Delete(id) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "task not found"})
			return
		}
		json.NewEncoder(w).Encode(map[string]bool{"deleted": true})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}
