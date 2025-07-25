package service

import (
	"context"
	"github.com/a2sh3r/gophermart/internal/accrual"
	"github.com/a2sh3r/gophermart/internal/apperrors"
	"github.com/a2sh3r/gophermart/internal/logger"
	"github.com/a2sh3r/gophermart/internal/models"
	"github.com/a2sh3r/gophermart/internal/repository"
	"github.com/a2sh3r/gophermart/internal/utils"
	"go.uber.org/zap"
	"time"
)

const (
	StatusNew       = "NEW"
	StatusProcessed = "PROCESSED"
)

type OrderService interface {
	UploadOrder(ctx context.Context, number string, userID int64) error
	GetUserOrders(ctx context.Context, userID int64) ([]models.Order, error)
}

type orderService struct {
	repo          repository.OrderRepository
	balanceRepo   repository.BalanceRepository
	accrualClient accrual.ClientInterface
}

func NewOrderService(repo repository.OrderRepository, balanceRepo repository.BalanceRepository, accrualClient accrual.ClientInterface) OrderService {
	return &orderService{
		repo:          repo,
		balanceRepo:   balanceRepo,
		accrualClient: accrualClient,
	}
}

func (s *orderService) UploadOrder(ctx context.Context, number string, userID int64) error {
	if !utils.IsValidLuhn(number) {
		return apperrors.ErrInvalidOrderNumber
	}

	ownerID, err := s.repo.GetOrderOwner(ctx, number)
	if err != nil {
		return err
	}

	switch {
	case ownerID == userID:
		return apperrors.ErrOrderExistsSameUser
	case ownerID != 0 && ownerID != userID:
		return apperrors.ErrOrderExistsOtherUser
	}

	logger.Log.Info("trying to get order status from accrual system", zap.String("order", number))
	accrualResp, statusCode, err := s.accrualClient.GetOrderStatus(ctx, number)
	if err != nil {
		logger.Log.Warn("accrual service error", zap.Error(err), zap.Int("statusCode", statusCode))
	} else {
		logger.Log.Info("accrual service response received", zap.Any("response", accrualResp))
	}

	status := StatusNew
	var accrualSum *float64

	if accrualResp != nil {
		status = string(accrualResp.Status)
		accrualSum = accrualResp.Accrual
	}

	order := &models.Order{
		Number:     number,
		Status:     status,
		Accrual:    accrualSum,
		UploadedAt: time.Now(),
		UserID:     userID,
	}

	if err := s.repo.SaveOrder(ctx, order); err != nil {
		return err
	}

	if status == StatusProcessed && accrualSum != nil && *accrualSum > 0 {
		if err := s.balanceRepo.IncreaseUserBalance(ctx, userID, *accrualSum); err != nil {
			logger.Log.Error("failed to increase user balance", zap.Error(err), zap.Int64("userID", userID), zap.Float64("accrual", *accrualSum))
			return err
		}
		logger.Log.Info("user balance increased", zap.Int64("userID", userID), zap.Float64("accrual", *accrualSum))
	}

	return nil
}

func (s *orderService) GetUserOrders(ctx context.Context, userID int64) ([]models.Order, error) {
	return s.repo.GetOrdersByUser(ctx, userID)
}
