package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/a2sh3r/gophermart/internal/apperrors"
	"github.com/a2sh3r/gophermart/internal/mocks/repository_mocks"
	"github.com/a2sh3r/gophermart/internal/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func validOrder() string {
	return "79927398713"
}

func invalidOrder() string {
	return "1234567890"
}

func TestBalanceService_Withdraw(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	tests := []struct {
		name           string
		userID         int64
		withdrawalReq  models.WithdrawalRequest
		mockGetBalance func(m *repository_mocks.MockBalanceRepository)
		mockWithdraw   func(m *repository_mocks.MockBalanceRepository)
		wantErr        error
	}{
		{
			name:   "успешный вывод средств",
			userID: 1,
			withdrawalReq: models.WithdrawalRequest{
				Order: validOrder(),
				Sum:   50,
			},
			mockGetBalance: func(m *repository_mocks.MockBalanceRepository) {
				m.EXPECT().GetBalance(ctx, int64(1)).Return(models.Balance{Current: 100}, nil).Times(1)
			},
			mockWithdraw: func(m *repository_mocks.MockBalanceRepository) {
				m.EXPECT().Withdraw(ctx, gomock.AssignableToTypeOf(models.Withdrawal{})).DoAndReturn(
					func(_ context.Context, w models.Withdrawal) error {
						assert.Equal(t, int64(1), w.UserID)
						assert.Equal(t, validOrder(), w.Order)
						assert.Equal(t, float64(50), w.Sum)
						assert.WithinDuration(t, time.Now(), w.Processed, time.Second)
						return nil
					}).Times(1)
			},
			wantErr: nil,
		},
		{
			name:           "невалидный номер заказа",
			userID:         1,
			withdrawalReq:  models.WithdrawalRequest{Order: invalidOrder(), Sum: 50},
			mockGetBalance: func(m *repository_mocks.MockBalanceRepository) {},
			mockWithdraw:   func(m *repository_mocks.MockBalanceRepository) {},
			wantErr:        apperrors.ErrInvalidOrderNumber,
		},
		{
			name:   "ошибка получения баланса",
			userID: 2,
			withdrawalReq: models.WithdrawalRequest{
				Order: validOrder(),
				Sum:   50,
			},
			mockGetBalance: func(m *repository_mocks.MockBalanceRepository) {
				m.EXPECT().GetBalance(ctx, int64(2)).Return(models.Balance{}, errors.New("db error")).Times(1)
			},
			mockWithdraw: func(m *repository_mocks.MockBalanceRepository) {},
			wantErr:      errors.New("db error"),
		},
		{
			name:   "недостаточно средств",
			userID: 3,
			withdrawalReq: models.WithdrawalRequest{
				Order: validOrder(),
				Sum:   150,
			},
			mockGetBalance: func(m *repository_mocks.MockBalanceRepository) {
				m.EXPECT().GetBalance(ctx, int64(3)).Return(models.Balance{Current: 100}, nil).Times(1)
			},
			mockWithdraw: func(m *repository_mocks.MockBalanceRepository) {},
			wantErr:      apperrors.ErrInsufficientFunds,
		},
		{
			name:           "некорректная сумма для вывода (<=0)",
			userID:         4,
			withdrawalReq:  models.WithdrawalRequest{Order: validOrder(), Sum: 0},
			mockGetBalance: func(m *repository_mocks.MockBalanceRepository) {},
			mockWithdraw:   func(m *repository_mocks.MockBalanceRepository) {},
			wantErr:        apperrors.ErrInvalidWithdrawalSum,
		},
		{
			name:   "ошибка записи вывода",
			userID: 5,
			withdrawalReq: models.WithdrawalRequest{
				Order: validOrder(),
				Sum:   30,
			},
			mockGetBalance: func(m *repository_mocks.MockBalanceRepository) {
				m.EXPECT().GetBalance(ctx, int64(5)).Return(models.Balance{Current: 50}, nil).Times(1)
			},
			mockWithdraw: func(m *repository_mocks.MockBalanceRepository) {
				m.EXPECT().Withdraw(ctx, gomock.AssignableToTypeOf(models.Withdrawal{})).Return(errors.New("write error")).Times(1)
			},
			wantErr: errors.New("write error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := repository_mocks.NewMockBalanceRepository(ctrl)
			tt.mockGetBalance(mockRepo)
			tt.mockWithdraw(mockRepo)

			svc := NewBalanceService(mockRepo)
			err := svc.Withdraw(ctx, tt.userID, tt.withdrawalReq)

			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error())
			}
		})
	}
}
