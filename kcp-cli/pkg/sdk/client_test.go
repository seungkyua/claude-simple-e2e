package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	c := NewClient("http://localhost:8080")
	if c.baseURL != "http://localhost:8080" {
		t.Errorf("baseURL = %s, want http://localhost:8080", c.baseURL)
	}
	if c.maxRetries != 3 {
		t.Errorf("maxRetries = %d, want 3", c.maxRetries)
	}
}

func TestClientWithOptions(t *testing.T) {
	c := NewClient("http://localhost",
		WithToken("test-token"),
		WithUserAgent("test-agent"),
		WithMaxRetries(5),
	)
	if c.token != "test-token" {
		t.Errorf("token = %s, want test-token", c.token)
	}
	if c.userAgent != "test-agent" {
		t.Errorf("userAgent = %s, want test-agent", c.userAgent)
	}
	if c.maxRetries != 5 {
		t.Errorf("maxRetries = %d, want 5", c.maxRetries)
	}
}

func TestClientGet(t *testing.T) {
	// 테스트 서버 생성
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Authorization header missing or incorrect")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	c := NewClient(server.URL, WithToken("test-token"), WithMaxRetries(0))
	var result map[string]string
	err := c.Get("/test", &result)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("status = %s, want ok", result["status"])
	}
}

func TestClientAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: ErrorDetail{
				Code:    "NOT_FOUND",
				Message: "리소스를 찾을 수 없습니다",
				Status:  404,
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, WithMaxRetries(0))
	err := c.Get("/missing", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode)
	}
	if apiErr.Code != "NOT_FOUND" {
		t.Errorf("Code = %s, want NOT_FOUND", apiErr.Code)
	}
}

func TestClientRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error: ErrorDetail{Code: "SERVICE_UNAVAILABLE", Message: "서비스 이용 불가", Status: 503},
			})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	c := NewClient(server.URL, WithMaxRetries(3))
	var result map[string]string
	err := c.Get("/retry", &result)
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}
