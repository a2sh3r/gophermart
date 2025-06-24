package handlers

import (
	"bytes"
	"errors"
	"github.com/a2sh3r/gophermart/internal/apperrors"
	"github.com/a2sh3r/gophermart/internal/mocks/service_mocks"
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
	h := &Handler{userService: mockUserService}

	tests := []struct {
		name           string
		body           string
		mockSetup      func()
		wantStatusCode int
	}{
		{
			name: "success",
			body: `{"login":"test","password":"password"}`,
			mockSetup: func() {
				mockUserService.EXPECT().Register(gomock.Any(), "test", "password").Return(nil)
				mockUserService.EXPECT().GetUserByLogin(gomock.Any(), "test").Return(&models.User{ID: 1, Login: "test"}, nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "user already exists",
			body: `{"login":"test","password":"password"}`,
			mockSetup: func() {
				mockUserService.EXPECT().Register(gomock.Any(), "test", "password").Return(apperrors.ErrUserAlreadyExists)
			},
			wantStatusCode: http.StatusConflict,
		},
		{
			name:           "invalid json",
			body:           `{"login":"test"`,
			mockSetup:      func() {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "service error",
			body: `{"login":"test","password":"password"}`,
			mockSetup: func() {
				mockUserService.EXPECT().Register(gomock.Any(), "test", "password").Return(errors.New("fail"))
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:           "empty login",
			body:           `{"login":"","password":"password"}`,
			mockSetup:      func() {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "empty password",
			body:           `{"login":"test","password":""}`,
			mockSetup:      func() {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "missing login",
			body:           `{"password":"password"}`,
			mockSetup:      func() {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "missing password",
			body:           `{"login":"test"}`,
			mockSetup:      func() {},
			wantStatusCode: http.StatusBadRequest,
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

func TestHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUserService := service_mocks.NewMockUserService(ctrl)
	h := &Handler{userService: mockUserService}

	tests := []struct {
		name           string
		body           string
		mockSetup      func()
		wantStatusCode int
		checkResponse  func(t *testing.T, resp *http.Response)
	}{
		{
			name: "success",
			body: `{"login":"test","password":"password"}`,
			mockSetup: func() {
				mockUserService.EXPECT().Authenticate(gomock.Any(), "test", "password").Return(nil)
				mockUserService.EXPECT().GetUserByLogin(gomock.Any(), "test").Return(&models.User{ID: 1, Login: "test"}, nil)
			},
			wantStatusCode: http.StatusOK,
			checkResponse: func(t *testing.T, resp *http.Response) {
				if resp.Header.Get("Authorization") == "" {
					t.Error("expected Authorization header")
				}
			},
		},
		{
			name: "invalid credentials",
			body: `{"login":"test","password":"wrong"}`,
			mockSetup: func() {
				mockUserService.EXPECT().Authenticate(gomock.Any(), "test", "wrong").Return(apperrors.ErrInvalidCredentials)
			},
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "invalid json",
			body:           `{"login":"test"`,
			mockSetup:      func() {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "service error",
			body: `{"login":"test","password":"password"}`,
			mockSetup: func() {
				mockUserService.EXPECT().Authenticate(gomock.Any(), "test", "password").Return(errors.New("fail"))
			},
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "empty login",
			body:           `{"login":"","password":"password"}`,
			mockSetup:      func() {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "empty password",
			body:           `{"login":"test","password":""}`,
			mockSetup:      func() {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "missing login",
			body:           `{"password":"password"}`,
			mockSetup:      func() {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "missing password",
			body:           `{"login":"test"}`,
			mockSetup:      func() {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "user not found",
			body: `{"login":"nonexistent","password":"password"}`,
			mockSetup: func() {
				mockUserService.EXPECT().Authenticate(gomock.Any(), "nonexistent", "password").Return(apperrors.ErrUserNotFound)
			},
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "very long login",
			body:           `{"login":"` + string(make([]byte, 1000)) + `","password":"password"}`,
			mockSetup:      func() {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "very long password",
			body:           `{"login":"test","password":"` + string(make([]byte, 1000)) + `"}`,
			mockSetup:      func() {},
			wantStatusCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBufferString(tt.body))
			w := httptest.NewRecorder()
			h.Login(w, req)
			resp := w.Result()
			if resp.StatusCode != tt.wantStatusCode {
				t.Errorf("got status %d, want %d", resp.StatusCode, tt.wantStatusCode)
			}
			if tt.checkResponse != nil {
				tt.checkResponse(t, resp)
			}
			err := resp.Body.Close()
			if err != nil {
				return
			}
		})
	}
}
