package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/a2sh3r/gophermart/internal/apperrors"
	"github.com/a2sh3r/gophermart/internal/models"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
}

type userRepo struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) CreateUser(ctx context.Context, user *models.User) error {
	existing, err := r.GetUserByLogin(ctx, user.Login)
	if err != nil && !errors.Is(err, apperrors.ErrUserNotFound) {
		return err
	}
	if existing != nil {
		return apperrors.ErrUserAlreadyExists
	}

	query := `INSERT INTO users (login, password_hash) VALUES ($1, $2)`
	_, err = r.db.ExecContext(ctx, query, user.Login, user.Password)
	return err
}

func (r *userRepo) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	query := `SELECT id, login, password_hash FROM users WHERE login=$1`
	row := r.db.QueryRowContext(ctx, query, login)

	var user models.User
	err := row.Scan(&user.ID, &user.Login, &user.Password)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}
