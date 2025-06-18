package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/a2sh3r/gophermart/internal/logger"
	"github.com/a2sh3r/gophermart/internal/models"
	"go.uber.org/zap"
)

type OrderRepository interface {
	SaveOrder(ctx context.Context, order *models.Order) error
	GetOrdersByUser(ctx context.Context, userID int64) ([]models.Order, error)
	GetOrderOwner(ctx context.Context, number string) (int64, error)
	GetUnprocessedOrders(ctx context.Context) ([]models.Order, error)
	UpdateOrderStatus(ctx context.Context, order *models.Order) error
}

type orderRepo struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepo{db: db}
}

func (r *orderRepo) SaveOrder(ctx context.Context, order *models.Order) error {
	query := `INSERT INTO orders (number, status, accrual, uploaded_at, user_id)
			  VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query,
		order.Number, order.Status, order.Accrual, order.UploadedAt, order.UserID)
	return err
}

func (r *orderRepo) GetOrdersByUser(ctx context.Context, userID int64) ([]models.Order, error) {
	query := `SELECT number, status, accrual, uploaded_at FROM orders
			  WHERE user_id=$1 ORDER BY uploaded_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		logger.Log.Error("failed to initiate query", zap.Error(err))
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			logger.Log.Error("failed to close rows", zap.Error(err))
		}
	}(rows)

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			logger.Log.Error("failed to scan order row", zap.Error(err))
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (r *orderRepo) GetOrderOwner(ctx context.Context, number string) (int64, error) {
	query := `SELECT user_id FROM orders WHERE number=$1`
	var userID int64
	err := r.db.QueryRowContext(ctx, query, number).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return userID, err
}

func (r *orderRepo) GetUnprocessedOrders(ctx context.Context) ([]models.Order, error) {
	query := `
			SELECT number, status, accrual, uploaded_at, user_id
			FROM orders
			WHERE status IN ('NEW', 'PROCESSING')
			ORDER BY uploaded_at
		`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			logger.Log.Error("failed to close rows", zap.Error(err))
		}
	}(rows)

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.Number, &o.Status, &o.Accrual, &o.UploadedAt, &o.UserID); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (r *orderRepo) UpdateOrderStatus(ctx context.Context, order *models.Order) error {
	query := `
		UPDATE orders
		SET status = $1, accrual = $2
		WHERE number = $3
	`
	_, err := r.db.ExecContext(ctx, query, order.Status, order.Accrual, order.Number)
	return err
}
