package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hex.EncodeToString(b[0:4]),
		hex.EncodeToString(b[4:6]),
		hex.EncodeToString(b[6:8]),
		hex.EncodeToString(b[8:10]),
		hex.EncodeToString(b[10:]))
}

const (
	statusProcessing = "processing"
	statusCompleted  = "completed"

	processingTTL = 5 * time.Minute
	completedTTL  = 24 * time.Hour
)

type CachedResponse struct {
	StatusCode int    `json:"status_code"`
	Body       []byte `json:"body"`
}

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(addr string) *RedisStore {
	return &RedisStore{
		client: redis.NewClient(&redis.Options{Addr: addr}),
	}
}

func (s *RedisStore) redisKey(token string) string {
	return "idempotency:" + token
}

func (s *RedisStore) StartProcessing(ctx context.Context, token string) (bool, error) {
	return s.client.SetNX(ctx, s.redisKey(token), statusProcessing, processingTTL).Result()
}

func (s *RedisStore) GetStatus(ctx context.Context, token string) (string, bool, error) {
	val, err := s.client.Get(ctx, s.redisKey(token)).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return val, true, nil
}

func (s *RedisStore) GetResult(ctx context.Context, token string) (*CachedResponse, error) {
	val, err := s.client.Get(ctx, s.redisKey(token)).Result()
	if err != nil {
		return nil, err
	}
	var cached CachedResponse
	if err := json.Unmarshal([]byte(val), &cached); err != nil {
		return nil, err
	}
	return &cached, nil
}

func (s *RedisStore) Finish(ctx context.Context, token string, statusCode int, body []byte) error {
	data, err := json.Marshal(CachedResponse{StatusCode: statusCode, Body: body})
	if err != nil {
		return err
	}
	return s.client.Set(ctx, s.redisKey(token), data, completedTTL).Err()
}

func IdempotencyMiddleware(store *RedisStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Idempotency-Key")
		if token == "" {
			http.Error(w, `{"error":"Idempotency-Key header required"}`, http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		reserved, err := store.StartProcessing(ctx, token)
		if err != nil {
			fmt.Printf("  [Redis] StartProcessing error: %v\n", err)
			http.Error(w, `{"error":"storage error"}`, http.StatusInternalServerError)
			return
		}

		if !reserved {
			rawVal, exists, err := store.GetStatus(ctx, token)
			if err != nil || !exists {
				http.Error(w, `{"error":"storage error"}`, http.StatusInternalServerError)
				return
			}

			if rawVal == statusProcessing {
				fmt.Printf("  [Middleware] Key %s... is processing -> 409 Conflict\n", token[:8])
				http.Error(w, `{"error":"Duplicate request in progress"}`, http.StatusConflict)
				return
			}

			cached, err := store.GetResult(ctx, token)
			if err != nil {
				http.Error(w, `{"error":"failed to read cached result"}`, http.StatusInternalServerError)
				return
			}
			fmt.Printf("  [Middleware] Key %s... already completed -> returning cached result\n", token[:8])
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Idempotent-Replayed", "true")
			w.WriteHeader(cached.StatusCode)
			w.Write(cached.Body)
			return
		}

		fmt.Printf("  [Middleware] New key %s... -> executing business logic\n", token[:8])
		recorder := httptest.NewRecorder()
		next.ServeHTTP(recorder, r)

		if err := store.Finish(ctx, token, recorder.Code, recorder.Body.Bytes()); err != nil {
			fmt.Printf("  [Redis] Finish error: %v\n", err)
		}
		fmt.Printf("  [Middleware] Key %s... completed, result saved to Redis (TTL 24h)\n", token[:8])

		for k, vals := range recorder.Header() {
			for _, v := range vals {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(recorder.Code)
		w.Write(recorder.Body.Bytes())
	})
}

func loanRepaymentHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("  [Business Logic] Processing started - simulating 2s heavy operation...")
	time.Sleep(2 * time.Second)

	transactionID := generateUUID()
	fmt.Printf("  [Business Logic] Payment complete, transaction_id: %s\n", transactionID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":         "paid",
		"amount":         1000,
		"transaction_id": transactionID,
	})
}

func main() {
	fmt.Println("  Task 2: Loan Repayment Idempotency (Redis)  \n")

	store := NewRedisStore("localhost:6379")

	ctx := context.Background()
	if err := store.client.Ping(ctx).Err(); err != nil {
		fmt.Printf("ERROR: Cannot connect to Redis at localhost:6379\n  %v\n", err)
		fmt.Println("\nStart Redis first:")
		fmt.Println("  Mac:   brew services start redis")
		fmt.Println("  Linux: sudo systemctl start redis")
		fmt.Println("  Check: redis-cli ping  (should print PONG)")
		return
	}
	fmt.Println("Redis connected. PONG received.\n")

	mux := http.NewServeMux()
	mux.HandleFunc("/loan/repay", loanRepaymentHandler)
	server := httptest.NewServer(IdempotencyMiddleware(store, mux))
	defer server.Close()

	fmt.Println("  Test 1: Missing Idempotency-Key  ")
	resp, _ := http.Post(server.URL+"/loan/repay", "application/json", nil)
	fmt.Printf("Response: %d  (expected 400)\n\n", resp.StatusCode)

	fmt.Println("  Test 2: Double-Click Attack (10 concurrent requests, same key)  ")
	fmt.Println("Business logic takes 2s. Requests arriving during those 2s -> 409 Conflict.\n")

	idempotencyKey := generateUUID()
	fmt.Printf("Idempotency-Key: %s\n\n", idempotencyKey)

	type result struct {
		goroutine  int
		statusCode int
		body       string
		replayed   bool
	}
	results := make(chan result, 10)
	var wg sync.WaitGroup

	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			time.Sleep(time.Duration(n*20) * time.Millisecond)

			req, _ := http.NewRequest("POST", server.URL+"/loan/repay", nil)
			req.Header.Set("Idempotency-Key", idempotencyKey)
			resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
			if err != nil {
				results <- result{goroutine: n, statusCode: -1, body: err.Error()}
				return
			}
			defer resp.Body.Close()
			var bodyMap map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&bodyMap)
			b, _ := json.Marshal(bodyMap)
			results <- result{
				goroutine:  n,
				statusCode: resp.StatusCode,
				body:       string(b),
				replayed:   resp.Header.Get("X-Idempotent-Replayed") == "true",
			}
		}(i)
	}

	go func() { wg.Wait(); close(results) }()

	var successCount, conflictCount, replayedCount int
	for r := range results {
		switch r.statusCode {
		case http.StatusOK:
			if r.replayed {
				replayedCount++
				fmt.Printf("Goroutine %2d: 200 OK  (REPLAYED CACHE)   %s\n", r.goroutine, r.body)
			} else {
				successCount++
				fmt.Printf("Goroutine %2d: 200 OK  (FIRST execution)  %s\n", r.goroutine, r.body)
			}
		case http.StatusConflict:
			conflictCount++
			fmt.Printf("Goroutine %2d: 409 Conflict  (duplicate in progress)\n", r.goroutine)
		default:
			fmt.Printf("Goroutine %2d: %d  %s\n", r.goroutine, r.statusCode, r.body)
		}
	}

	fmt.Printf("\n  Summary  \n")
	fmt.Printf("First execution (business logic ran):  %d\n", successCount)
	fmt.Printf("Conflicts (duplicate in progress):     %d\n", conflictCount)
	fmt.Printf("Replayed (cached result returned):     %d\n", replayedCount)

	fmt.Println("\n  Test 3: Late request after completion (should return cached 200)  ")
	req, _ := http.NewRequest("POST", server.URL+"/loan/repay", nil)
	req.Header.Set("Idempotency-Key", idempotencyKey)
	lateResp, _ := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	defer lateResp.Body.Close()
	var lateBody map[string]interface{}
	json.NewDecoder(lateResp.Body).Decode(&lateBody)
	fmt.Printf("Status:              %d\n", lateResp.StatusCode)
	fmt.Printf("Body:                %v\n", lateBody)
	fmt.Printf("X-Idempotent-Replayed: %s\n", lateResp.Header.Get("X-Idempotent-Replayed"))
	fmt.Println("\nBusiness logic was NOT re-executed. ✓")

	fmt.Println("\n  Test 4: Redis key verification  ")
	redisKey := "idempotency:" + idempotencyKey
	val, err := store.client.Get(ctx, redisKey).Result()
	if err != nil {
		fmt.Printf("Redis key not found: %v\n", err)
	} else {
		var cached CachedResponse
		json.Unmarshal([]byte(val), &cached)
		ttl, _ := store.client.TTL(ctx, redisKey).Result()
		fmt.Printf("Redis key exists:  idempotency:%s...\n", idempotencyKey[:8])
		fmt.Printf("Stored HTTP code:  %d\n", cached.StatusCode)
		fmt.Printf("Stored body:       %s\n", string(cached.Body))
		fmt.Printf("TTL remaining:     %v\n", ttl.Round(time.Second))
	}
}
