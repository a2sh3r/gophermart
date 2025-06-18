package service

import (
	"context"
	"github.com/a2sh3r/gophermart/internal/apperrors"
	"github.com/a2sh3r/gophermart/internal/models"
	"github.com/a2sh3r/gophermart/internal/repository"
	"github.com/a2sh3r/gophermart/internal/utils"
	"time"
)

type BalanceService interface {
	GetUserBalance(ctx context.Context, userID int64) (models.Balance, error)
	Withdraw(ctx context.Context, userID int64, withdrawal models.WithdrawalRequest) error
}

type balanceService struct {
	repo repository.BalanceRepository
}

func NewBalanceService(repo repository.BalanceRepository) BalanceService {
	return &balanceService{repo: repo}
}

func (s *balanceService) GetUserBalance(ctx context.Context, userID int64) (models.Balance, error) {
	return s.repo.GetBalance(ctx, userID)
}

func (s *balanceService) Withdraw(ctx context.Context, userID int64, withdrawalReq models.WithdrawalRequest) error {
	if !utils.IsValidLuhn(withdrawalReq.Order) {
		return apperrors.ErrInvalidOrderNumber
	}

	balance, err := s.repo.GetBalance(ctx, userID)
	if err != nil {
		return err
	}

	if withdrawalReq.Sum <= 0 {
		return apperrors.ErrInvalidWithdrawalSum
	}

	if balance.Current < withdrawalReq.Sum {
		return apperrors.ErrInsufficientFunds
	}

	withdrawal := models.Withdrawal{
		Order:     withdrawalReq.Order,
		Sum:       withdrawalReq.Sum,
		Processed: time.Now(),
		UserID:    userID,
	}

	return s.repo.Withdraw(ctx, withdrawal)
}
