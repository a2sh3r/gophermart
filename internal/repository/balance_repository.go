package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/a2sh3r/gophermart/internal/logger"
	"github.com/a2sh3r/gophermart/internal/models"
	"go.uber.org/zap"
)

type BalanceRepository interface {
	GetBalance(ctx context.Context, userID int64) (models.Balance, error)
	Withdraw(ctx context.Context, withdrawal models.Withdrawal) error
	GetWithdrawals(ctx context.Context, userID int64) ([]models.Withdrawal, error)
	IncreaseUserBalance(ctx context.Context, userID int64, accrual float64) error
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
		SELECT current_balance, withdrawn_balance FROM users WHERE id = $1
	`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&balance.Current, &balance.Withdrawn)

	if errors.Is(err, sql.ErrNoRows) {
		return models.Balance{Current: 0, Withdrawn: 0}, nil
	}
	if err != nil {
		logger.Log.Error("failed to get balance", zap.Error(err))
		return models.Balance{}, err
	}
	return balance, nil
}

func (r *balanceRepo) IncreaseUserBalance(ctx context.Context, userID int64, accrual float64) error {
	query := `
		UPDATE users
		SET current_balance = current_balance + $1
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, accrual, userID)
	return err
}

func (r *balanceRepo) Withdraw(ctx context.Context, withdrawal models.Withdrawal) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				logger.Log.Error("rollback error")
				return
			}
		}
	}()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO withdrawals (order_number, sum, processed_at, user_id)
		VALUES ($1, $2, $3, $4)
	`, withdrawal.Order, withdrawal.Sum, withdrawal.Processed, withdrawal.UserID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE users
		SET current_balance = current_balance - $1,
		    withdrawn_balance = withdrawn_balance + $1
		WHERE id = $2 AND current_balance >= $1
	`, withdrawal.Sum, withdrawal.UserID)
	if err != nil {
		return err
	}

	err = tx.Commit()
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
		w.UserID = userID
		withdrawals = append(withdrawals, w)
	}
	return withdrawals, nil
}
