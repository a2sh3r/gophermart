package handlers

import (
	"bytes"
	"errors"
	"github.com/a2sh3r/gophermart/internal/apperrors"
	service_mocks "github.com/a2sh3r/gophermart/internal/mocks/service_mocks"
	"github.com/a2sh3r/gophermart/internal/models"
	"github.com/golang/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := service_mocks.NewMockUserService(ctrl)
	h := &Handler{userService: mockUserService, secretKey: "test"}

	tests := []struct {
		name           string
		body           string
		mockSetup      func()
		wantStatusCode int
	}{
		{
			name: "success",
			body: `{"login":"user","password":"pass"}`,
			mockSetup: func() {
				mockUserService.EXPECT().Register(gomock.Any(), "user", "pass").Return(nil)
				mockUserService.EXPECT().GetUserByLogin(gomock.Any(), "user").Return(&models.User{ID: 1, Login: "user"}, nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "user already exists",
			body: `{"login":"user","password":"pass"}`,
			mockSetup: func() {
				mockUserService.EXPECT().Register(gomock.Any(), "user", "pass").Return(apperrors.ErrUserAlreadyExists)
			},
			wantStatusCode: http.StatusConflict,
		},
		{
			name:           "invalid json",
			body:           `{"login":""}`,
			mockSetup:      func() {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "service error",
			body: `{"login":"user","password":"pass"}`,
			mockSetup: func() {
				mockUserService.EXPECT().Register(gomock.Any(), "user", "pass").Return(errors.New("fail"))
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBufferString(tt.body))
			w := httptest.NewRecorder()
			h.Register(w, req)
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
