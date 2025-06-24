package service

import (
	"context"
	"errors"
	"github.com/a2sh3r/gophermart/internal/apperrors"

	"github.com/a2sh3r/gophermart/internal/models"
	"github.com/a2sh3r/gophermart/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, login, password string) error
	Authenticate(ctx context.Context, login, password string) error
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) Register(ctx context.Context, login, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &models.User{
		Login:    login,
		Password: string(hashedPassword),
	}

	err = s.repo.CreateUser(ctx, user)
	if errors.Is(err, apperrors.ErrUserAlreadyExists) {
		return err
	}
	return err
}

func (s *userService) Authenticate(ctx context.Context, login, password string) error {
	user, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		return apperrors.ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return apperrors.ErrInvalidCredentials
	}

	return nil
}

func (s *userService) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	return s.repo.GetUserByLogin(ctx, login)
}
