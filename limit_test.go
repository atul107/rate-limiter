package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(1 * time.Minute)
	rl.AddRule("/test", 2)

	handler := rl.MiddleWare((http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Test response"))
	})))

	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.Header.Set("client-key", "client-1")
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Errorf("Expected: %d, got: %d", http.StatusOK, w1.Code)
	}

	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("client-key", "client-1")
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("Expected: %d, got: %d", http.StatusOK, w2.Code)
	}

	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.Header.Set("client-key", "client-1")
	w3 := httptest.NewRecorder()
	handler.ServeHTTP(w3, req3)
	if w3.Code != http.StatusTooManyRequests {
		t.Errorf("Expected: %d, got: %d", http.StatusTooManyRequests, w3.Code)
	}
}
