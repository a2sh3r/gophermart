package service

import (
	"context"
	"errors"
	"testing"

	"github.com/a2sh3r/gophermart/internal/apperrors"
	"github.com/a2sh3r/gophermart/internal/mocks/repository_mocks"
	"github.com/a2sh3r/gophermart/internal/models"
	"github.com/golang/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestUserService_Register(t *testing.T) {
	tests := []struct {
		name        string
		login       string
		password    string
		mockSetup   func(m *repository_mocks.MockUserRepository)
		expectedErr error
	}{
		{
			name:     "успешная регистрация",
			login:    "user1",
			password: "password123",
			mockSetup: func(m *repository_mocks.MockUserRepository) {
				m.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name:     "пользователь уже существует",
			login:    "user2",
			password: "password123",
			mockSetup: func(m *repository_mocks.MockUserRepository) {
				m.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(apperrors.ErrUserAlreadyExists)
			},
			expectedErr: apperrors.ErrUserAlreadyExists,
		},
		{
			name:     "неизвестная ошибка создания",
			login:    "user3",
			password: "password123",
			mockSetup: func(m *repository_mocks.MockUserRepository) {
				m.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(errors.New("db fail"))
			},
			expectedErr: errors.New("db fail"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := repository_mocks.NewMockUserRepository(ctrl)
			tt.mockSetup(repo)

			service := NewUserService(repo)
			err := service.Register(context.Background(), tt.login, tt.password)

			if tt.expectedErr != nil && err.Error() != tt.expectedErr.Error() {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
			if tt.expectedErr == nil && err != nil {
				t.Errorf("expected nil error, got %v", err)
			}
		})
	}
}

func TestUserService_Authenticate(t *testing.T) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name        string
		login       string
		password    string
		mockUser    *models.User
		mockErr     error
		expectedErr error
	}{
		{
			name:     "успешная аутентификация",
			login:    "user1",
			password: "password123",
			mockUser: &models.User{Login: "user1", Password: string(hashed)},
		},
		{
			name:        "неправильный пароль",
			login:       "user2",
			password:    "wrongpass",
			mockUser:    &models.User{Login: "user2", Password: string(hashed)},
			expectedErr: apperrors.ErrInvalidCredentials,
		},
		{
			name:        "пользователь не найден",
			login:       "user3",
			password:    "any",
			mockErr:     errors.New("not found"),
			expectedErr: apperrors.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := repository_mocks.NewMockUserRepository(ctrl)
			repo.EXPECT().GetUserByLogin(gomock.Any(), tt.login).Return(tt.mockUser, tt.mockErr)

			service := NewUserService(repo)
			err := service.Authenticate(context.Background(), tt.login, tt.password)

			if tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
			if tt.expectedErr == nil && err != nil {
				t.Errorf("expected nil error, got %v", err)
			}
		})
	}
}

func TestUserService_GetUserByLogin(t *testing.T) {
	expectedUser := &models.User{Login: "user1", Password: "hashed"}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repository_mocks.NewMockUserRepository(ctrl)
	repo.EXPECT().GetUserByLogin(gomock.Any(), "user1").Return(expectedUser, nil)

	service := NewUserService(repo)

	user, err := service.GetUserByLogin(context.Background(), "user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Login != expectedUser.Login {
		t.Errorf("expected login %s, got %s", expectedUser.Login, user.Login)
	}
}
