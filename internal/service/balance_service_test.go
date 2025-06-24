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

func TestBalanceService_GetUserBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockBalanceRepo := repository_mocks.NewMockBalanceRepository(ctrl)
	s := &balanceService{repo: mockBalanceRepo}

	tests := []struct {
		name      string
		userID    int64
		mockSetup func()
		want      models.Balance
		wantErr   bool
	}{
		{
			name:   "success",
			userID: 1,
			mockSetup: func() {
				mockBalanceRepo.EXPECT().GetBalance(gomock.Any(), int64(1)).Return(models.Balance{Current: 100.50, Withdrawn: 25.00}, nil)
			},
			want:    models.Balance{Current: 100.50, Withdrawn: 25.00},
			wantErr: false,
		},
		{
			name:   "repository error",
			userID: 1,
			mockSetup: func() {
				mockBalanceRepo.EXPECT().GetBalance(gomock.Any(), int64(1)).Return(models.Balance{}, errors.New("database error"))
			},
			want:    models.Balance{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			got, err := s.GetUserBalance(context.Background(), tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserBalance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetUserBalance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBalanceService_GetWithdrawals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockBalanceRepo := repository_mocks.NewMockBalanceRepository(ctrl)
	s := &balanceService{repo: mockBalanceRepo}

	tests := []struct {
		name      string
		userID    int64
		mockSetup func()
		want      []models.Withdrawal
		wantErr   bool
	}{
		{
			name:   "success with withdrawals",
			userID: 1,
			mockSetup: func() {
				withdrawals := []models.Withdrawal{
					{Order: "12345678903", Sum: 100.50, UserID: 1},
					{Order: "98765432109", Sum: 200.00, UserID: 1},
				}
				mockBalanceRepo.EXPECT().GetWithdrawals(gomock.Any(), int64(1)).Return(withdrawals, nil)
			},
			want: []models.Withdrawal{
				{Order: "12345678903", Sum: 100.50, UserID: 1},
				{Order: "98765432109", Sum: 200.00, UserID: 1},
			},
			wantErr: false,
		},
		{
			name:   "success no withdrawals",
			userID: 1,
			mockSetup: func() {
				mockBalanceRepo.EXPECT().GetWithdrawals(gomock.Any(), int64(1)).Return([]models.Withdrawal{}, nil)
			},
			want:    []models.Withdrawal{},
			wantErr: false,
		},
		{
			name:   "repository error",
			userID: 1,
			mockSetup: func() {
				mockBalanceRepo.EXPECT().GetWithdrawals(gomock.Any(), int64(1)).Return(nil, errors.New("database error"))
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			got, err := s.GetWithdrawals(context.Background(), tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetWithdrawals() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("GetWithdrawals() length = %v, want %v", len(got), len(tt.want))
				return
			}
			for i, withdrawal := range got {
				if withdrawal != tt.want[i] {
					t.Errorf("GetWithdrawals()[%d] = %v, want %v", i, withdrawal, tt.want[i])
				}
			}
		})
	}
}
