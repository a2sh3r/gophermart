package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/a2sh3r/gophermart/internal/apperrors"
	"github.com/a2sh3r/gophermart/internal/middleware"
	service_mocks "github.com/a2sh3r/gophermart/internal/mocks/service_mocks"
	"github.com/a2sh3r/gophermart/internal/models"
	"github.com/golang/mock/gomock"
	"io"
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
			err := resp.Body.Close()
			if err != nil {
				return
			}
		})
	}
}

func TestHandler_Withdraw(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockBalanceService := service_mocks.NewMockBalanceService(ctrl)
	h := &Handler{balanceService: mockBalanceService}

	tests := []struct {
		name           string
		userID         int64
		body           string
		mockSetup      func()
		wantStatusCode int
	}{
		{
			name:   "success",
			userID: 1,
			body:   `{"order":"12345678903","sum":100.50}`,
			mockSetup: func() {
				mockBalanceService.EXPECT().Withdraw(gomock.Any(), int64(1), models.WithdrawalRequest{Order: "12345678903", Sum: 100.50}).Return(nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:   "invalid order number",
			userID: 1,
			body:   `{"order":"123","sum":100.50}`,
			mockSetup: func() {
				mockBalanceService.EXPECT().Withdraw(gomock.Any(), int64(1), models.WithdrawalRequest{Order: "123", Sum: 100.50}).Return(apperrors.ErrInvalidOrderNumber)
			},
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name:   "insufficient funds",
			userID: 1,
			body:   `{"order":"12345678903","sum":1000.00}`,
			mockSetup: func() {
				mockBalanceService.EXPECT().Withdraw(gomock.Any(), int64(1), models.WithdrawalRequest{Order: "12345678903", Sum: 1000.00}).Return(apperrors.ErrInsufficientFunds)
			},
			wantStatusCode: http.StatusPaymentRequired,
		},
		{
			name:   "invalid withdrawal sum",
			userID: 1,
			body:   `{"order":"12345678903","sum":-50.00}`,
			mockSetup: func() {
				mockBalanceService.EXPECT().Withdraw(gomock.Any(), int64(1), models.WithdrawalRequest{Order: "12345678903", Sum: -50.00}).Return(apperrors.ErrInvalidWithdrawalSum)
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:   "service error",
			userID: 1,
			body:   `{"order":"12345678903","sum":100.50}`,
			mockSetup: func() {
				mockBalanceService.EXPECT().Withdraw(gomock.Any(), int64(1), models.WithdrawalRequest{Order: "12345678903", Sum: 100.50}).Return(errors.New("fail"))
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBufferString(tt.body))
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			h.Withdraw(w, req)
			resp := w.Result()
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					fmt.Printf("%v", err)
				}
			}(resp.Body)
			if resp.StatusCode != tt.wantStatusCode {
				t.Errorf("got status %d, want %d", resp.StatusCode, tt.wantStatusCode)
			}
		})
	}
}

func TestHandler_GetWithdrawals(t *testing.T) {
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
			name:   "success with withdrawals",
			userID: 1,
			mockSetup: func() {
				withdrawals := []models.Withdrawal{
					{Order: "12345678903", Sum: 100.50, UserID: 1},
					{Order: "98765432109", Sum: 200.00, UserID: 1},
				}
				mockBalanceService.EXPECT().GetWithdrawals(gomock.Any(), int64(1)).Return(withdrawals, nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:   "success no withdrawals",
			userID: 1,
			mockSetup: func() {
				mockBalanceService.EXPECT().GetWithdrawals(gomock.Any(), int64(1)).Return([]models.Withdrawal{}, nil)
			},
			wantStatusCode: http.StatusNoContent,
		},
		{
			name:   "service error",
			userID: 1,
			mockSetup: func() {
				mockBalanceService.EXPECT().GetWithdrawals(gomock.Any(), int64(1)).Return(nil, errors.New("fail"))
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			h.GetWithdrawals(w, req)
			resp := w.Result()
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					fmt.Printf("%v", err)
				}
			}(resp.Body)
			if resp.StatusCode != tt.wantStatusCode {
				t.Errorf("got status %d, want %d", resp.StatusCode, tt.wantStatusCode)
			}
		})
	}
}
