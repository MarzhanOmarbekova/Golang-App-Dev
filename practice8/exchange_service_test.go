package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestExchangeService(server *httptest.Server) *ExchangeService {
	svc := NewExchangeService(server.URL)
	svc.Client = server.Client()
	return svc
}

func TestGetRate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/convert", r.URL.Path)
		assert.Equal(t, "USD", r.URL.Query().Get("from"))
		assert.Equal(t, "EUR", r.URL.Query().Get("to"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"base":"USD","target":"EUR","rate":0.92}`))
	}))
	defer server.Close()

	svc := newTestExchangeService(server)
	rate, err := svc.GetRate("USD", "EUR")

	require.NoError(t, err)
	assert.InDelta(t, 0.92, rate, 0.0001)
}

func TestGetRate_APIBusinessError_404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"invalid currency pair"}`))
	}))
	defer server.Close()

	svc := newTestExchangeService(server)
	_, err := svc.GetRate("USD", "XYZ")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid currency pair")
}

func TestGetRate_APIBusinessError_400(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid currency pair"}`))
	}))
	defer server.Close()

	svc := newTestExchangeService(server)
	_, err := svc.GetRate("", "EUR")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid currency pair")
}

func TestGetRate_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`Internal Server Error`))
	}))
	defer server.Close()

	svc := newTestExchangeService(server)
	_, err := svc.GetRate("USD", "EUR")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "decode error")
}

func TestGetRate_Timeout(t *testing.T) {
	done := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-done:
		case <-r.Context().Done():
		}
	}))
	defer func() {
		close(done)
		server.Close()
	}()

	svc := newTestExchangeService(server)
	svc.Client = &http.Client{Timeout: 50 * time.Millisecond}

	_, err := svc.GetRate("USD", "EUR")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "network error")
}

func TestGetRate_ServerPanic_500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
	}))
	defer server.Close()

	svc := newTestExchangeService(server)
	_, err := svc.GetRate("USD", "EUR")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "internal server error")
}

func TestGetRate_EmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := newTestExchangeService(server)
	_, err := svc.GetRate("USD", "EUR")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "decode error")
}
