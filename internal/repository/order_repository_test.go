package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/a2sh3r/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/lib/pq"
)

func float64Ptr(v float64) *float64 {
	return &v
}

func setupOrderTestData(t *testing.T, db *sql.DB) {
	_, err := db.Exec(`TRUNCATE orders, users RESTART IDENTITY CASCADE`)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO users (id, login, password_hash, created_at, current_balance, withdrawn_balance) 
		VALUES 
		(1, 'user1', 'hash1', now(), 100, 0),
		(2, 'user2', 'hash2', now(), 200, 50)
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO orders (number, status, accrual, uploaded_at, user_id) VALUES
		('1234567890', 'NEW', NULL, now() - interval '1 hour', 1),
		('0987654321', 'PROCESSING', NULL, now() - interval '30 minutes', 1),
		('1111111111', 'PROCESSED', 100, now() - interval '15 minutes', 1),
		('2222222222', 'INVALID', NULL, now() - interval '10 minutes', 2),
		('3333333333', 'PROCESSED', 200, now(), 2)
	`)
	require.NoError(t, err)
}

func TestOrderRepo_SaveOrder(t *testing.T) {
	r := NewOrderRepository(testDB)
	ctx := context.Background()

	tests := []struct {
		name      string
		order     *models.Order
		wantErr   bool
		setupFunc func()
	}{
		{
			name: "save new order",
			order: &models.Order{
				Number:     "9999999999",
				Status:     "NEW",
				Accrual:    nil,
				UploadedAt: time.Now(),
				UserID:     1,
			},
			wantErr: false,
			setupFunc: func() {
				setupOrderTestData(t, testDB)
			},
		},
		{
			name: "save order with accrual",
			order: &models.Order{
				Number:     "8888888888",
				Status:     "PROCESSED",
				Accrual:    float64Ptr(150),
				UploadedAt: time.Now(),
				UserID:     2,
			},
			wantErr: false,
			setupFunc: func() {
				setupOrderTestData(t, testDB)
			},
		},
		{
			name: "save order with invalid status",
			order: &models.Order{
				Number:     "7777777777",
				Status:     "INVALID",
				Accrual:    nil,
				UploadedAt: time.Now(),
				UserID:     1,
			},
			wantErr: false,
			setupFunc: func() {
				setupOrderTestData(t, testDB)
			},
		},
		{
			name: "save duplicate order",
			order: &models.Order{
				Number:     "1234567890",
				Status:     "NEW",
				Accrual:    nil,
				UploadedAt: time.Now(),
				UserID:     1,
			},
			wantErr: true,
			setupFunc: func() {
				setupOrderTestData(t, testDB)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			err := r.SaveOrder(ctx, tt.order)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			var count int
			err = testDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM orders WHERE number = $1`, tt.order.Number).Scan(&count)
			assert.NoError(t, err)
			assert.Equal(t, 1, count)
		})
	}
}

func TestOrderRepo_GetOrdersByUser(t *testing.T) {
	r := NewOrderRepository(testDB)
	ctx := context.Background()

	tests := []struct {
		name      string
		userID    int64
		wantCount int
		wantErr   bool
		setupFunc func()
	}{
		{
			name:      "user with multiple orders",
			userID:    1,
			wantCount: 3,
			wantErr:   false,
			setupFunc: func() {
				setupOrderTestData(t, testDB)
			},
		},
		{
			name:      "user with orders",
			userID:    2,
			wantCount: 2,
			wantErr:   false,
			setupFunc: func() {
				setupOrderTestData(t, testDB)
			},
		},
		{
			name:      "user without orders",
			userID:    999,
			wantCount: 0,
			wantErr:   false,
			setupFunc: func() {
				setupOrderTestData(t, testDB)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			orders, err := r.GetOrdersByUser(ctx, tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, orders, tt.wantCount)

			if len(orders) > 1 {
				for i := 0; i < len(orders)-1; i++ {
					assert.True(t, orders[i].UploadedAt.After(orders[i+1].UploadedAt) || orders[i].UploadedAt.Equal(orders[i+1].UploadedAt))
				}
			}

			for _, order := range orders {
				assert.NotEmpty(t, order.Number)
				assert.NotEmpty(t, order.Status)
			}
		})
	}
}

func TestOrderRepo_GetOrderOwner(t *testing.T) {
	r := NewOrderRepository(testDB)
	ctx := context.Background()

	tests := []struct {
		name      string
		number    string
		wantUser  int64
		wantErr   bool
		setupFunc func()
	}{
		{
			name:     "existing order",
			number:   "1234567890",
			wantUser: 1,
			wantErr:  false,
			setupFunc: func() {
				setupOrderTestData(t, testDB)
			},
		},
		{
			name:     "another existing order",
			number:   "2222222222",
			wantUser: 2,
			wantErr:  false,
			setupFunc: func() {
				setupOrderTestData(t, testDB)
			},
		},
		{
			name:     "non-existing order",
			number:   "9999999999",
			wantUser: 0,
			wantErr:  false,
			setupFunc: func() {
				setupOrderTestData(t, testDB)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			userID, err := r.GetOrderOwner(ctx, tt.number)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantUser, userID)
		})
	}
}

func TestOrderRepo_GetUnprocessedOrders(t *testing.T) {
	r := NewOrderRepository(testDB)
	ctx := context.Background()

	tests := []struct {
		name      string
		wantCount int
		wantErr   bool
		setupFunc func()
	}{
		{
			name:      "get unprocessed orders",
			wantCount: 2,
			wantErr:   false,
			setupFunc: func() {
				setupOrderTestData(t, testDB)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			orders, err := r.GetUnprocessedOrders(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, orders, tt.wantCount)

			for _, order := range orders {
				assert.Contains(t, []string{"NEW", "PROCESSING"}, order.Status)
				assert.NotEmpty(t, order.Number)
				assert.NotZero(t, order.UserID)
			}

			if len(orders) > 1 {
				for i := 0; i < len(orders)-1; i++ {
					assert.True(t, orders[i].UploadedAt.Before(orders[i+1].UploadedAt) || orders[i].UploadedAt.Equal(orders[i+1].UploadedAt))
				}
			}
		})
	}
}

func TestOrderRepo_UpdateOrderStatus(t *testing.T) {
	r := NewOrderRepository(testDB)
	ctx := context.Background()

	tests := []struct {
		name      string
		order     *models.Order
		wantErr   bool
		setupFunc func()
	}{
		{
			name: "update to processed",
			order: &models.Order{
				Number:  "1234567890",
				Status:  "PROCESSED",
				Accrual: float64Ptr(100),
			},
			wantErr: false,
			setupFunc: func() {
				setupOrderTestData(t, testDB)
			},
		},
		{
			name: "update to invalid",
			order: &models.Order{
				Number:  "0987654321",
				Status:  "INVALID",
				Accrual: nil,
			},
			wantErr: false,
			setupFunc: func() {
				setupOrderTestData(t, testDB)
			},
		},
		{
			name: "update to invalid with existing order",
			order: &models.Order{
				Number:  "1111111111",
				Status:  "INVALID",
				Accrual: nil,
			},
			wantErr: false,
			setupFunc: func() {
				setupOrderTestData(t, testDB)
			},
		},
		{
			name: "update non-existing order",
			order: &models.Order{
				Number:  "9999999999",
				Status:  "PROCESSED",
				Accrual: float64Ptr(50),
			},
			wantErr: false,
			setupFunc: func() {
				setupOrderTestData(t, testDB)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			err := r.UpdateOrderStatus(ctx, tt.order)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			var status string
			var accrual sql.NullFloat64
			err = testDB.QueryRowContext(ctx, `SELECT status, accrual FROM orders WHERE number = $1`, tt.order.Number).Scan(&status, &accrual)
			if err != nil && err != sql.ErrNoRows {
				assert.NoError(t, err)
			}

			if err != sql.ErrNoRows {
				assert.Equal(t, tt.order.Status, status)
				if tt.order.Accrual != nil {
					assert.True(t, accrual.Valid)
					assert.Equal(t, *tt.order.Accrual, accrual.Float64)
				} else {
					assert.False(t, accrual.Valid)
				}
			}
		})
	}
}
