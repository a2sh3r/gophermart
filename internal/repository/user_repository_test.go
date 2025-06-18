package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/a2sh3r/gophermart/internal/apperrors"
	"github.com/a2sh3r/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserTestData(t *testing.T, db *sql.DB) {
	_, err := db.Exec(`TRUNCATE users RESTART IDENTITY CASCADE`)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO users (login, password_hash, created_at, current_balance, withdrawn_balance) 
		VALUES 
		('user1', 'hash1', now(), 100, 0),
		('user2', 'hash2', now(), 200, 50)
	`)
	require.NoError(t, err)
}

func TestUserRepo_CreateUser(t *testing.T) {
	r := NewUserRepository(testDB)
	ctx := context.Background()

	tests := []struct {
		name      string
		user      *models.User
		wantErr   bool
		setupFunc func()
	}{
		{
			name: "create new user",
			user: &models.User{
				Login:    "newuser",
				Password: "newhash",
			},
			wantErr: false,
			setupFunc: func() {
				setupUserTestData(t, testDB)
			},
		},
		{
			name: "create user with existing login",
			user: &models.User{
				Login:    "user1",
				Password: "differenthash",
			},
			wantErr: true,
			setupFunc: func() {
				setupUserTestData(t, testDB)
			},
		},
		{
			name: "create another new user",
			user: &models.User{
				Login:    "user3",
				Password: "hash3",
			},
			wantErr: false,
			setupFunc: func() {
				setupUserTestData(t, testDB)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			err := r.CreateUser(ctx, tt.user)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, apperrors.ErrUserAlreadyExists)
				return
			}
			assert.NoError(t, err)

			var count int
			err = testDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE login = $1`, tt.user.Login).Scan(&count)
			assert.NoError(t, err)
			assert.Equal(t, 1, count)
		})
	}
}

func TestUserRepo_GetUserByLogin(t *testing.T) {
	r := NewUserRepository(testDB)
	ctx := context.Background()

	tests := []struct {
		name      string
		login     string
		wantErr   bool
		setupFunc func()
	}{
		{
			name:    "existing user",
			login:   "user1",
			wantErr: false,
			setupFunc: func() {
				setupUserTestData(t, testDB)
			},
		},
		{
			name:    "another existing user",
			login:   "user2",
			wantErr: false,
			setupFunc: func() {
				setupUserTestData(t, testDB)
			},
		},
		{
			name:    "non-existing user",
			login:   "nonexistent",
			wantErr: true,
			setupFunc: func() {
				setupUserTestData(t, testDB)
			},
		},
		{
			name:    "empty login",
			login:   "",
			wantErr: true,
			setupFunc: func() {
				setupUserTestData(t, testDB)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			user, err := r.GetUserByLogin(ctx, tt.login)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.login != "" {
					assert.ErrorIs(t, err, apperrors.ErrUserNotFound)
				}
				assert.Nil(t, user)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, user)
			assert.Equal(t, tt.login, user.Login)
			assert.NotEmpty(t, user.Password)
		})
	}
}
