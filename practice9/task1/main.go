package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"
)

type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

type PaymentClient struct {
	httpClient *http.Client
	config     RetryConfig
}

func NewPaymentClient(cfg RetryConfig) *PaymentClient {
	return &PaymentClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		config:     cfg,
	}
}

func IsRetryable(resp *http.Response, err error) bool {
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return true
		}
		return true
	}
	if resp != nil {
		switch resp.StatusCode {
		case 429, 500, 502, 503, 504:
			return true
		case 401, 403, 404, 400:
			return false
		}
	}
	return false
}

func CalculateBackoff(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
	backoff := baseDelay * time.Duration(math.Pow(2, float64(attempt)))
	if backoff > maxDelay {
		backoff = maxDelay
	}
	jitter := time.Duration(rand.Int63n(int64(backoff) + 1))
	return jitter
}

func (c *PaymentClient) ExecutePayment(ctx context.Context, url string) (*http.Response, error) {
	var (
		resp *http.Response
		err  error
	)

	for attempt := 0; attempt < c.config.MaxRetries; attempt++ {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
		}

		req, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
		if reqErr != nil {
			return nil, fmt.Errorf("failed to create request: %w", reqErr)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err = c.httpClient.Do(req)

		if err == nil && resp.StatusCode == http.StatusOK {
			fmt.Printf("Attempt %d: Success!\n", attempt+1)
			return resp, nil
		}

		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
			resp.Body.Close()
		}

		if !IsRetryable(resp, err) {
			if err != nil {
				return nil, fmt.Errorf("non-retryable error: %w", err)
			}
			return nil, fmt.Errorf("non-retryable status code: %d", statusCode)
		}

		if attempt == c.config.MaxRetries-1 {
			if err != nil {
				return nil, fmt.Errorf("all %d attempts failed, last error: %w", c.config.MaxRetries, err)
			}
			return nil, fmt.Errorf("all %d attempts failed, last status: %d", c.config.MaxRetries, statusCode)
		}

		waitTime := CalculateBackoff(attempt, c.config.BaseDelay, c.config.MaxDelay)

		if err != nil {
			fmt.Printf("Attempt %d failed (error: %v): waiting ~%v...\n", attempt+1, err, waitTime)
		} else {
			fmt.Printf("Attempt %d failed (status %d): waiting ~%v...\n", attempt+1, statusCode, waitTime)
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled while waiting: %w", ctx.Err())
		case <-time.After(waitTime):
		}
	}

	return nil, errors.New("retry loop exited unexpectedly")
}

func main() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("    Task 1: Resilient HTTP Client    \n")

	var requestCount int32

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		if count <= 3 {
			fmt.Printf("  [Server] Request #%d -> 503 Service Unavailable\n", count)
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error": "service unavailable"}`))
		} else {
			fmt.Printf("  [Server] Request #%d -> 200 OK\n", count)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "success"}`))
		}
	}))
	defer testServer.Close()

	client := NewPaymentClient(RetryConfig{
		MaxRetries: 5,
		BaseDelay:  500 * time.Millisecond,
		MaxDelay:   5 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("Starting payment execution...")
	fmt.Println("Server will return 503 for first 3 requests, then 200 OK.\n")

	resp, err := client.ExecutePayment(ctx, testServer.URL+"/pay")
	if err != nil {
		fmt.Printf("\nFinal result: FAILED - %v\n", err)
	} else {
		defer resp.Body.Close()
		fmt.Printf("\nFinal result: SUCCESS (status %d)\n", resp.StatusCode)
	}

	fmt.Println("\n   IsRetryable examples   ")
	fmt.Printf("429 is retryable: %v\n", IsRetryable(&http.Response{StatusCode: 429}, nil))
	fmt.Printf("503 is retryable: %v\n", IsRetryable(&http.Response{StatusCode: 503}, nil))
	fmt.Printf("404 is retryable: %v\n", IsRetryable(&http.Response{StatusCode: 404}, nil))
	fmt.Printf("401 is retryable: %v\n", IsRetryable(&http.Response{StatusCode: 401}, nil))
	fmt.Printf("network error is retryable: %v\n", IsRetryable(nil, errors.New("connection refused")))
}
