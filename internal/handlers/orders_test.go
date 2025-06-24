package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/a2sh3r/gophermart/internal/apperrors"
	"github.com/a2sh3r/gophermart/internal/middleware"
	"github.com/a2sh3r/gophermart/internal/mocks/service_mocks"
	"github.com/a2sh3r/gophermart/internal/models"
	"github.com/golang/mock/gomock"
	"io"
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
		{
			name:   "order already exists",
			body:   "12345678903",
			userID: 1,
			mockSetup: func() {
				mockOrderService.EXPECT().UploadOrder(gomock.Any(), "12345678903", int64(1)).Return(apperrors.ErrOrderExistsSameUser)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:   "order already exists for another user",
			body:   "12345678903",
			userID: 1,
			mockSetup: func() {
				mockOrderService.EXPECT().UploadOrder(gomock.Any(), "12345678903", int64(1)).Return(apperrors.ErrOrderExistsOtherUser)
			},
			wantStatusCode: http.StatusConflict,
		},
		{
			name:           "invalid order number (fails Luhn check)",
			body:           "123",
			userID:         1,
			mockSetup:      func() {},
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name:           "empty body",
			body:           "",
			userID:         1,
			mockSetup:      func() {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "whitespace only body",
			body:           "   ",
			userID:         1,
			mockSetup:      func() {},
			wantStatusCode: http.StatusUnprocessableEntity,
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
			err := resp.Body.Close()
			if err != nil {
				return
			}
		})
	}
}

func TestHandler_GetOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockOrderService := service_mocks.NewMockOrderService(ctrl)
	h := &Handler{orderService: mockOrderService}

	tests := []struct {
		name           string
		userID         int64
		mockSetup      func()
		wantStatusCode int
	}{
		{
			name:   "success with orders",
			userID: 1,
			mockSetup: func() {
				orders := []models.Order{
					{Number: "12345678903", Status: "NEW", UserID: 1},
					{Number: "98765432109", Status: "PROCESSED", UserID: 1},
				}
				mockOrderService.EXPECT().GetUserOrders(gomock.Any(), int64(1)).Return(orders, nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:   "success no orders",
			userID: 1,
			mockSetup: func() {
				mockOrderService.EXPECT().GetUserOrders(gomock.Any(), int64(1)).Return([]models.Order{}, nil)
			},
			wantStatusCode: http.StatusNoContent,
		},
		{
			name:   "service error",
			userID: 1,
			mockSetup: func() {
				mockOrderService.EXPECT().GetUserOrders(gomock.Any(), int64(1)).Return(nil, errors.New("fail"))
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			h.GetOrders(w, req)
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
