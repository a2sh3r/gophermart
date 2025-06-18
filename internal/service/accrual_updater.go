package service

import (
	"context"
	"github.com/a2sh3r/gophermart/internal/accrual"
	"github.com/a2sh3r/gophermart/internal/logger"
	"github.com/a2sh3r/gophermart/internal/repository"
	"go.uber.org/zap"
	"time"
)

type AccrualUpdater struct {
	repo          repository.OrderRepository
	balanceRepo   repository.BalanceRepository
	accrualClient accrual.ClientInterface
	pollInterval  time.Duration
}

func NewAccrualUpdater(repo repository.OrderRepository, balanceRepo repository.BalanceRepository, client accrual.ClientInterface, interval time.Duration) *AccrualUpdater {
	return &AccrualUpdater{
		repo:          repo,
		balanceRepo:   balanceRepo,
		accrualClient: client,
		pollInterval:  interval,
	}
}

func (u *AccrualUpdater) Run(ctx context.Context) {
	ticker := time.NewTicker(u.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			u.checkAndUpdateOrders(ctx)
		}
	}
}

func (u *AccrualUpdater) checkAndUpdateOrders(ctx context.Context) {
	orders, err := u.repo.GetUnprocessedOrders(ctx)
	if err != nil {
		logger.Log.Error("failed to get unprocessed orders", zap.Error(err))
		return
	}

	for _, order := range orders {
		resp, _, err := u.accrualClient.GetOrderStatus(ctx, order.Number)
		if err != nil {
			logger.Log.Warn("failed to get accrual status", zap.String("order", order.Number), zap.Error(err))
			continue
		}

		if resp == nil {
			continue
		}

		order.Status = string(resp.Status)
		if order.Status == string(accrual.StatusRegistered) {
			order.Status = StatusNew
		}

		order.Accrual = resp.Accrual

		if err := u.repo.UpdateOrderStatus(ctx, &order); err != nil {
			logger.Log.Error("failed to update order", zap.String("order", order.Number), zap.Error(err))
		}

		if resp.Status == accrual.StatusProcessed && resp.Accrual != nil {
			if err := u.balanceRepo.IncreaseUserBalance(ctx, order.UserID, *resp.Accrual); err != nil {
				logger.Log.Error("failed to increase balance", zap.Int64("user", order.UserID), zap.Error(err))
			}
		}
	}
}
