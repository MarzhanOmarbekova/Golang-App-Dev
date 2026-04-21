package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

type CachedResponse struct {
	StatusCode int
	Body       []byte
	Completed  bool
}

type MemoryStore struct {
	mu   sync.Mutex
	data map[string]*CachedResponse
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{data: make(map[string]*CachedResponse)}
}

func (m *MemoryStore) Get(key string) (*CachedResponse, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	resp, exists := m.data[key]
	return resp, exists
}

func (m *MemoryStore) StartProcessing(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.data[key]; exists {
		return false
	}
	m.data[key] = &CachedResponse{Completed: false}
	return true
}

func (m *MemoryStore) Finish(key string, status int, body []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if resp, exists := m.data[key]; exists {
		resp.StatusCode = status
		resp.Body = body
		resp.Completed = true
	} else {
		m.data[key] = &CachedResponse{StatusCode: status, Body: body, Completed: true}
	}
}

func IdempotencyMiddleware(store *MemoryStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			http.Error(w, "Idempotency-Key header required", http.StatusBadRequest)
			return
		}
		if cached, exists := store.Get(key); exists {
			if cached.Completed {
				w.WriteHeader(cached.StatusCode)
				w.Write(cached.Body)
			} else {
				http.Error(w, "Duplicate request in progress", http.StatusConflict)
			}
			return
		}
		if !store.StartProcessing(key) {
			if cached, exists := store.Get(key); exists && cached.Completed {
				w.WriteHeader(cached.StatusCode)
				w.Write(cached.Body)
			} else {
				http.Error(w, "Duplicate request in progress", http.StatusConflict)
			}
			return
		}
		recorder := httptest.NewRecorder()
		next.ServeHTTP(recorder, r)
		store.Finish(key, recorder.Code, recorder.Body.Bytes())
		for k, vals := range recorder.Header() {
			for _, v := range vals {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(recorder.Code)
		w.Write(recorder.Body.Bytes())
	})
}

func paymentHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("  [Business Logic] Executing payment...")
	time.Sleep(500 * time.Millisecond)
	resp := map[string]interface{}{
		"status": "paid",
		"amount": 1000,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func main() {
	fmt.Println("   Part 2 Theory Example: Idempotency Middleware   \n")

	store := NewMemoryStore()
	mux := http.NewServeMux()
	mux.HandleFunc("/pay", paymentHandler)
	handler := IdempotencyMiddleware(store, mux)
	server := httptest.NewServer(handler)
	defer server.Close()

	client := &http.Client{}
	idempotencyKey := "demo-key-12345"

	fmt.Println("   Sending request without Idempotency-Key   ")
	req, _ := http.NewRequest("POST", server.URL+"/pay", nil)
	resp, _ := client.Do(req)
	fmt.Printf("Response: %d\n", resp.StatusCode)

	fmt.Println("\n   First request with key   ")
	req, _ = http.NewRequest("POST", server.URL+"/pay", nil)
	req.Header.Set("Idempotency-Key", idempotencyKey)
	resp, _ = client.Do(req)
	fmt.Printf("Response: %d\n", resp.StatusCode)

	fmt.Println("\n   Duplicate request with same key   ")
	req, _ = http.NewRequest("POST", server.URL+"/pay", nil)
	req.Header.Set("Idempotency-Key", idempotencyKey)
	resp, _ = client.Do(req)
	fmt.Printf("Response: %d (cached, no re-execution)\n", resp.StatusCode)

	fmt.Println("\nDone! Idempotency middleware works correctly.")
}
