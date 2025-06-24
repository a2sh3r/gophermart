package handlers

import (
	"github.com/a2sh3r/gophermart/internal/mocks/service_mocks"
	"github.com/golang/mock/gomock"
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
		err := resp.Body.Close()
		if err != nil {
			return
		}
	}
}

func TestNewHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := service_mocks.NewMockUserService(ctrl)
	mockOrderService := service_mocks.NewMockOrderService(ctrl)
	mockBalanceService := service_mocks.NewMockBalanceService(ctrl)

	h := NewHandler(mockUserService, mockOrderService, mockBalanceService, "test-secret")

	if h == nil {
		t.Fatal("NewHandler returned nil")
	}

	if h.userService == nil {
		t.Error("userService is nil")
	}
	if h.orderService == nil {
		t.Error("orderService is nil")
	}
	if h.balanceService == nil {
		t.Error("balanceService is nil")
	}
}
