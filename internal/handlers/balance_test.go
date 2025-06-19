package handlers

import (
	"context"
	"errors"
	"github.com/a2sh3r/gophermart/internal/middleware"
	service_mocks "github.com/a2sh3r/gophermart/internal/mocks/service_mocks"
	"github.com/a2sh3r/gophermart/internal/models"
	"github.com/golang/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_GetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockBalanceService := service_mocks.NewMockBalanceService(ctrl)
	h := &Handler{balanceService: mockBalanceService}

	tests := []struct {
		name           string
		userID         int64
		mockSetup      func()
		wantStatusCode int
	}{
		{
			name:   "success",
			userID: 1,
			mockSetup: func() {
				mockBalanceService.EXPECT().GetUserBalance(gomock.Any(), int64(1)).Return(models.Balance{Current: 100, Withdrawn: 0}, nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:   "service error",
			userID: 1,
			mockSetup: func() {
				mockBalanceService.EXPECT().GetUserBalance(gomock.Any(), int64(1)).Return(models.Balance{}, errors.New("fail"))
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			h.GetBalance(w, req)
			resp := w.Result()
			if resp.StatusCode != tt.wantStatusCode {
				t.Errorf("got status %d, want %d", resp.StatusCode, tt.wantStatusCode)
			}
		})
	}
}
