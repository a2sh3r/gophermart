package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/a2sh3r/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/lib/pq"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error
	testDB, err = sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/gophermart?sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer func(testDB *sql.DB) {
		err := testDB.Close()
		if err != nil {
			fmt.Printf("close db error")
		}
	}(testDB)

	_, err = testDB.Exec(`TRUNCATE orders, withdrawals, users RESTART IDENTITY CASCADE`)
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func setupTestData(t *testing.T, db *sql.DB) {
	_, err := db.Exec(`TRUNCATE orders, withdrawals, users RESTART IDENTITY CASCADE`)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO users (id, login, password_hash, created_at, current_balance, withdrawn_balance) 
		VALUES 
		(1, 'testuser1', 'fakehash1', now(), 100, 50),
		(2, 'testuser2', 'fakehash2', now(), 0, 0),
		(3, 'testuser3', 'fakehash3', now(), 200, 100)
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO orders (number, user_id, accrual, status, uploaded_at) VALUES
		('1234567890', 1, 100, 'PROCESSED', now()),
		('0987654321', 1, 50, 'PROCESSED', now()),
		('1111111111', 2, 0, 'INVALID', now()),
		('2222222222', 3, 200, 'PROCESSED', now())
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO withdrawals (order_number, sum, processed_at, user_id) VALUES
		('withdraw1', 20, now() - interval '1 day', 1),
		('withdraw2', 30, now(), 1),
		('withdraw3', 50, now() - interval '2 days', 3),
		('withdraw4', 50, now() - interval '1 hour', 3)
	`)
	require.NoError(t, err)
}

func TestBalanceRepo_GetBalance(t *testing.T) {
	r := NewBalanceRepository(testDB)
	ctx := context.Background()

	setupTestData(t, testDB)

	tests := []struct {
		name    string
		userID  int64
		want    models.Balance
		wantErr bool
	}{
		{
			name:   "user with positive balance",
			userID: 1,
			want: models.Balance{
				Current:   100,
				Withdrawn: 50,
			},
			wantErr: false,
		},
		{
			name:   "user with zero balance",
			userID: 2,
			want: models.Balance{
				Current:   0,
				Withdrawn: 0,
			},
			wantErr: false,
		},
		{
			name:   "user with high balance",
			userID: 3,
			want: models.Balance{
				Current:   200,
				Withdrawn: 100,
			},
			wantErr: false,
		},
		{
			name:    "non-existing user",
			userID:  999,
			want:    models.Balance{Current: 0, Withdrawn: 0},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r.GetBalance(ctx, tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.Current, got.Current)
				assert.Equal(t, tt.want.Withdrawn, got.Withdrawn)
			}
		})
	}
}

func TestBalanceRepo_IncreaseUserBalance(t *testing.T) {
	r := NewBalanceRepository(testDB)
	ctx := context.Background()

	tests := []struct {
		name      string
		userID    int64
		amount    float64
		wantErr   bool
		setupFunc func()
	}{
		{
			name:    "increase balance for existing user",
			userID:  1,
			amount:  50,
			wantErr: false,
			setupFunc: func() {
				setupTestData(t, testDB)
			},
		},
		{
			name:    "increase balance by zero",
			userID:  1,
			amount:  0,
			wantErr: false,
			setupFunc: func() {
				setupTestData(t, testDB)
			},
		},
		{
			name:    "increase balance by negative amount",
			userID:  1,
			amount:  -10,
			wantErr: false,
			setupFunc: func() {
				setupTestData(t, testDB)
			},
		},
		{
			name:    "increase balance for non-existing user",
			userID:  999,
			amount:  100,
			wantErr: false,
			setupFunc: func() {
				setupTestData(t, testDB)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			var initialBalance float64
			err := testDB.QueryRowContext(ctx, `SELECT current_balance FROM users WHERE id = $1`, tt.userID).Scan(&initialBalance)
			if err != nil && err != sql.ErrNoRows {
				require.NoError(t, err)
			}
			if err == sql.ErrNoRows {
				initialBalance = 0
			}

			err = r.IncreaseUserBalance(ctx, tt.userID, tt.amount)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			var newBalance float64
			err = testDB.QueryRowContext(ctx, `SELECT current_balance FROM users WHERE id = $1`, tt.userID).Scan(&newBalance)
			if err != nil && err != sql.ErrNoRows {
				assert.NoError(t, err)
			}
			if err == sql.ErrNoRows {
				newBalance = 0
			}

			if tt.userID == 999 {
				assert.Equal(t, 0.0, newBalance)
			} else {
				expectedBalance := initialBalance + tt.amount
				assert.Equal(t, expectedBalance, newBalance)
			}
		})
	}
}

func TestBalanceRepo_Withdraw(t *testing.T) {
	r := NewBalanceRepository(testDB)
	ctx := context.Background()

	tests := []struct {
		name       string
		withdrawal models.Withdrawal
		wantErr    bool
		setupFunc  func()
	}{
		{
			name: "successful withdrawal",
			withdrawal: models.Withdrawal{
				Order:     "test-order-1",
				Sum:       30,
				Processed: time.Now(),
				UserID:    1,
			},
			wantErr: false,
			setupFunc: func() {
				setupTestData(t, testDB)
			},
		},
		{
			name: "withdrawal with insufficient funds",
			withdrawal: models.Withdrawal{
				Order:     "test-order-2",
				Sum:       1000,
				Processed: time.Now(),
				UserID:    1,
			},
			wantErr: false,
			setupFunc: func() {
				setupTestData(t, testDB)
			},
		},
		{
			name: "withdrawal for user with zero balance",
			withdrawal: models.Withdrawal{
				Order:     "test-order-3",
				Sum:       10,
				Processed: time.Now(),
				UserID:    2,
			},
			wantErr: false,
			setupFunc: func() {
				setupTestData(t, testDB)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			var initialCurrent, initialWithdrawn float64
			err := testDB.QueryRowContext(ctx, `SELECT current_balance, withdrawn_balance FROM users WHERE id = $1`, tt.withdrawal.UserID).Scan(&initialCurrent, &initialWithdrawn)
			if err != nil && err != sql.ErrNoRows {
				require.NoError(t, err)
			}

			err = r.Withdraw(ctx, tt.withdrawal)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			var newCurrent, newWithdrawn float64
			err = testDB.QueryRowContext(ctx, `SELECT current_balance, withdrawn_balance FROM users WHERE id = $1`, tt.withdrawal.UserID).Scan(&newCurrent, &newWithdrawn)
			if err != nil && err != sql.ErrNoRows {
				assert.NoError(t, err)
			}

			var count int
			err = testDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM withdrawals WHERE order_number = $1 AND user_id = $2`, tt.withdrawal.Order, tt.withdrawal.UserID).Scan(&count)
			assert.NoError(t, err)
			assert.Equal(t, 1, count)

			if tt.name == "successful withdrawal" {
				expectedCurrent := initialCurrent - tt.withdrawal.Sum
				expectedWithdrawn := initialWithdrawn + tt.withdrawal.Sum

				assert.Equal(t, expectedCurrent, newCurrent)
				assert.Equal(t, expectedWithdrawn, newWithdrawn)
			} else {
				assert.Equal(t, initialCurrent, newCurrent)
				assert.Equal(t, initialWithdrawn, newWithdrawn)
			}
		})
	}
}

func TestBalanceRepo_GetWithdrawals(t *testing.T) {
	r := NewBalanceRepository(testDB)
	ctx := context.Background()

	tests := []struct {
		name      string
		userID    int64
		wantCount int
		wantErr   bool
		setupFunc func()
	}{
		{
			name:      "user with withdrawals",
			userID:    1,
			wantCount: 2,
			wantErr:   false,
			setupFunc: func() {
				setupTestData(t, testDB)
			},
		},
		{
			name:      "user with multiple withdrawals",
			userID:    3,
			wantCount: 2,
			wantErr:   false,
			setupFunc: func() {
				setupTestData(t, testDB)
			},
		},
		{
			name:      "user without withdrawals",
			userID:    2,
			wantCount: 0,
			wantErr:   false,
			setupFunc: func() {
				setupTestData(t, testDB)
			},
		},
		{
			name:      "non-existing user",
			userID:    999,
			wantCount: 0,
			wantErr:   false,
			setupFunc: func() {
				setupTestData(t, testDB)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			withdrawals, err := r.GetWithdrawals(ctx, tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, withdrawals, tt.wantCount)

			if len(withdrawals) > 1 {
				for i := 0; i < len(withdrawals)-1; i++ {
					assert.True(t, withdrawals[i].Processed.After(withdrawals[i+1].Processed) || withdrawals[i].Processed.Equal(withdrawals[i+1].Processed))
				}
			}

			for _, w := range withdrawals {
				assert.Equal(t, tt.userID, w.UserID)
			}
		})
	}
}
