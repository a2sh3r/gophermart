package handlers

import (
	"context"
	"errors"
	"github.com/a2sh3r/gophermart/internal/middleware"
	service_mocks "github.com/a2sh3r/gophermart/internal/mocks/service_mocks"
	"github.com/golang/mock/gomock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_UploadOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockOrderService := service_mocks.NewMockOrderService(ctrl)
	h := &Handler{orderService: mockOrderService}

	tests := []struct {
		name           string
		body           string
		userID         int64
		mockSetup      func()
		wantStatusCode int
	}{
		{
			name:   "success",
			body:   "12345678903",
			userID: 1,
			mockSetup: func() {
				mockOrderService.EXPECT().UploadOrder(gomock.Any(), "12345678903", int64(1)).Return(nil)
			},
			wantStatusCode: http.StatusAccepted,
		},
		{
			name:           "invalid format",
			body:           "abc",
			userID:         1,
			mockSetup:      func() {},
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name:   "service error",
			body:   "12345678903",
			userID: 1,
			mockSetup: func() {
				mockOrderService.EXPECT().UploadOrder(gomock.Any(), "12345678903", int64(1)).Return(errors.New("fail"))
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader(tt.body))
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			h.UploadOrder(w, req)
			resp := w.Result()
			if resp.StatusCode != tt.wantStatusCode {
				t.Errorf("got status %d, want %d", resp.StatusCode, tt.wantStatusCode)
			}
		})
	}
}
