package service

import (
	"context"
	"github.com/a2sh3r/gophermart/internal/apperrors"
	"github.com/a2sh3r/gophermart/internal/constants"
	"github.com/a2sh3r/gophermart/internal/models"
	"github.com/a2sh3r/gophermart/internal/repository"
	"github.com/a2sh3r/gophermart/internal/utils"
	"time"
)

type OrderService interface {
	UploadOrder(ctx context.Context, number string, userID int64) error
	GetUserOrders(ctx context.Context, userID int64) ([]models.Order, error)
}

type orderService struct {
	repo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) OrderService {
	return &orderService{repo: repo}
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

	order := &models.Order{
		Number:     number,
		Status:     constants.StatusNew,
		Accrual:    nil,
		UploadedAt: time.Now(),
		UserID:     userID,
	}

	return s.repo.SaveOrder(ctx, order)
}

func (s *orderService) GetUserOrders(ctx context.Context, userID int64) ([]models.Order, error) {
	return s.repo.GetOrdersByUser(ctx, userID)
}
