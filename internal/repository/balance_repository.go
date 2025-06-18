package repository

import (
	"context"
	"database/sql"
	"github.com/a2sh3r/gophermart/internal/logger"
	"github.com/a2sh3r/gophermart/internal/models"
	"go.uber.org/zap"
)

type BalanceRepository interface {
	GetBalance(ctx context.Context, userID int64) (models.Balance, error)
	Withdraw(ctx context.Context, withdrawal models.Withdrawal) error
	GetWithdrawals(ctx context.Context, userID int64) ([]models.Withdrawal, error)
}

type balanceRepo struct {
	db *sql.DB
}

func NewBalanceRepository(db *sql.DB) BalanceRepository {
	return &balanceRepo{db: db}
}

func (r *balanceRepo) GetBalance(ctx context.Context, userID int64) (models.Balance, error) {
	var balance models.Balance
	query := `
		SELECT
			COALESCE((SELECT SUM(accrual) FROM orders WHERE user_id = $1), 0)
			- COALESCE((SELECT SUM("sum") FROM withdrawals WHERE user_id = $1), 0) AS current,
			COALESCE((SELECT SUM("sum") FROM withdrawals WHERE user_id = $1), 0) AS withdrawn
	`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		logger.Log.Error("failed to get balance", zap.Error(err))
		return models.Balance{}, err
	}

	logger.Log.Info("balance loaded", zap.Any("balance", balance), zap.Int64("userID", userID))

	return balance, nil
}

func (r *balanceRepo) Withdraw(ctx context.Context, withdrawal models.Withdrawal) error {
	query := `
		INSERT INTO withdrawals (order_number, sum, processed_at, user_id)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecContext(ctx, query,
		withdrawal.Order, withdrawal.Sum, withdrawal.Processed, withdrawal.UserID)
	if err != nil {
		logger.Log.Error("failed to insert withdrawal", zap.Error(err))
	}
	return err
}

func (r *balanceRepo) GetWithdrawals(ctx context.Context, userID int64) ([]models.Withdrawal, error) {
	query := `
		SELECT order_number, sum, processed_at FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		logger.Log.Error("failed to query withdrawals", zap.Error(err))
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			logger.Log.Error("failed to close rows", zap.Error(err))
		}
	}(rows)

	var withdrawals []models.Withdrawal
	for rows.Next() {
		var w models.Withdrawal
		if err := rows.Scan(&w.Order, &w.Sum, &w.Processed); err != nil {
			logger.Log.Error("failed to scan withdrawal", zap.Error(err))
			return nil, err
		}
		withdrawals = append(withdrawals, w)
	}
	return withdrawals, nil
}
