package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouter_Routes(t *testing.T) {
	handler := &Handler{}
	router := NewRouter(handler, "testsecret")

	tests := []struct {
		method string
		path   string
		status int
	}{
		{"GET", "/api/user/orders", http.StatusUnauthorized},
		{"POST", "/api/user/register", http.StatusBadRequest},
		{"POST", "/api/user/login", http.StatusBadRequest},
		{"GET", "/notfound", http.StatusNotFound},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		resp := w.Result()
		if resp.StatusCode != tt.status {
			t.Errorf("%s %s: got %d, want %d", tt.method, tt.path, resp.StatusCode, tt.status)
		}
	}
}
