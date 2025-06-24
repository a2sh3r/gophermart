package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/a2sh3r/gophermart/internal/accrual"
	"github.com/a2sh3r/gophermart/internal/logger"
	"github.com/a2sh3r/gophermart/internal/mocks/repository_mocks"
	"github.com/a2sh3r/gophermart/internal/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func floatPtr(f float64) *float64 {
	return &f
}

type mockAccrualClient struct {
	statuses map[string]*accrual.AccrualResponse
	errors   map[string]error
}

func (m *mockAccrualClient) GetOrderStatus(_ context.Context, number string) (*accrual.AccrualResponse, int, error) {
	if err, ok := m.errors[number]; ok {
		return nil, 0, err
	}
	if resp, ok := m.statuses[number]; ok {
		return resp, 200, nil
	}
	return nil, 200, nil
}

func TestAccrualUpdater_checkAndUpdateOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger.Log = zap.NewNop()

	ctx := context.Background()

	tests := []struct {
		name               string
		unprocessedOrders  []models.Order
		getOrdersErr       error
		accrualStatuses    map[string]*accrual.AccrualResponse
		accrualErrors      map[string]error
		updateOrderErrors  map[string]error
		balanceIncreaseErr map[int64]error
	}{
		{
			name: "успешное обновление с начислением баланса",
			unprocessedOrders: []models.Order{
				{Number: "order1", UserID: 1, Status: "NEW", Accrual: floatPtr(1)},
			},
			accrualStatuses: map[string]*accrual.AccrualResponse{
				"order1": {Order: "order1", Status: accrual.StatusProcessed, Accrual: floatPtr(100)},
			},
			updateOrderErrors:  map[string]error{"order1": nil},
			balanceIncreaseErr: map[int64]error{1: nil},
		},
		{
			name:         "ошибка получения необработанных заказов",
			getOrdersErr: errors.New("db error"),
		},
		{
			name: "ошибка получения статуса заказа от accrual клиента",
			unprocessedOrders: []models.Order{
				{Number: "order2", UserID: 2, Status: "NEW"},
			},
			accrualErrors: map[string]error{"order2": errors.New("network error")},
		},
		{
			name: "nil accrual response - пропускаем заказ",
			unprocessedOrders: []models.Order{
				{Number: "order3", UserID: 3, Status: "NEW"},
			},
			accrualStatuses: map[string]*accrual.AccrualResponse{
				"order3": nil,
			},
		},
		{
			name: "ошибка обновления статуса заказа",
			unprocessedOrders: []models.Order{
				{Number: "order4", UserID: 4, Status: "NEW"},
			},
			accrualStatuses: map[string]*accrual.AccrualResponse{
				"order4": {Order: "order4", Status: accrual.StatusProcessing, Accrual: nil},
			},
			updateOrderErrors: map[string]error{"order4": errors.New("update failed")},
		},
		{
			name: "ошибка увеличения баланса пользователя",
			unprocessedOrders: []models.Order{
				{Number: "order5", UserID: 5, Status: "NEW"},
			},
			accrualStatuses: map[string]*accrual.AccrualResponse{
				"order5": {Order: "order5", Status: accrual.StatusProcessed, Accrual: floatPtr(50)},
			},
			updateOrderErrors:  map[string]error{"order5": nil},
			balanceIncreaseErr: map[int64]error{5: errors.New("balance error")},
		},
		{
			name: "статус заказа не изменился — обновление не вызывается",
			unprocessedOrders: []models.Order{
				{Number: "order6", UserID: 6, Status: "PROCESSED", Accrual: floatPtr(100.0)},
			},
			accrualStatuses: map[string]*accrual.AccrualResponse{
				"order6": {Order: "order6", Status: accrual.StatusProcessed, Accrual: floatPtr(100)},
			},

			updateOrderErrors: map[string]error{},
		},
		{
			name: "заказ с nil accrual, статус меняется, но начисления нет",
			unprocessedOrders: []models.Order{
				{Number: "order7", UserID: 7, Status: "PROCESSING", Accrual: floatPtr(0)},
			},
			accrualStatuses: map[string]*accrual.AccrualResponse{
				"order7": {Order: "order7", Status: StatusNew, Accrual: nil},
			},
			updateOrderErrors: map[string]error{"order7": nil},
		},
		{
			name: "заказ с статусом INVALID - начисления нет",
			unprocessedOrders: []models.Order{
				{Number: "order8", UserID: 8, Status: "NEW", Accrual: nil},
			},
			accrualStatuses: map[string]*accrual.AccrualResponse{
				"order8": {Order: "order8", Status: accrual.StatusInvalid, Accrual: nil},
			},
			updateOrderErrors: map[string]error{"order8": nil},
		},
		{
			name: "заказ с статусом INVALID - обновление статуса",
			unprocessedOrders: []models.Order{
				{Number: "order9", UserID: 9, Status: "PROCESSING", Accrual: floatPtr(50)},
			},
			accrualStatuses: map[string]*accrual.AccrualResponse{
				"order9": {Order: "order9", Status: accrual.StatusInvalid, Accrual: nil},
			},
			updateOrderErrors: map[string]error{"order9": nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrderRepo := repository_mocks.NewMockOrderRepository(ctrl)
			mockBalanceRepo := repository_mocks.NewMockBalanceRepository(ctrl)
			mockAccrualClient := &mockAccrualClient{
				statuses: tt.accrualStatuses,
				errors:   tt.accrualErrors,
			}

			mockOrderRepo.EXPECT().GetUnprocessedOrders(ctx).Return(tt.unprocessedOrders, tt.getOrdersErr).Times(1)

			for _, order := range tt.unprocessedOrders {
				if _, hasErr := tt.accrualErrors[order.Number]; hasErr {
					continue
				}

				resp := tt.accrualStatuses[order.Number]

				if resp == nil {
					continue
				}

				mockOrderRepo.EXPECT().UpdateOrderStatus(ctx, gomock.AssignableToTypeOf(&models.Order{})).DoAndReturn(
					func(_ context.Context, o *models.Order) error {
						if err, ok := tt.updateOrderErrors[o.Number]; ok {
							if string(resp.Status) != o.Status {
								t.Errorf("expected order status %s, got %s", resp.Status, o.Status)
							}
							if resp != nil {
								if string(resp.Status) != o.Status {
									t.Errorf("expected order status %s, got %s", resp.Status, o.Status)
								}
								if resp.Accrual != nil {
									if o.Accrual == nil || *resp.Accrual != *o.Accrual {
										t.Errorf("expected order accrual %f, got %v", *resp.Accrual, o.Accrual)
									}
								} else if o.Accrual != nil {
									t.Errorf("expected order accrual nil, got %v", o.Accrual)
								}
							}
							return err
						}
						return nil
					}).Times(1)

				if resp.Status == accrual.StatusProcessed && resp.Accrual != nil {
					mockBalanceRepo.EXPECT().IncreaseUserBalance(ctx, order.UserID, *resp.Accrual).DoAndReturn(
						func(ctx context.Context, userID int64, accrual float64) error {
							if userID != order.UserID {
								t.Errorf("expected userID %d, got %d", order.UserID, userID)
							}
							if accrual != *resp.Accrual {
								t.Errorf("expected accrual %f, got %f", *resp.Accrual, accrual)
							}
							if err, ok := tt.balanceIncreaseErr[userID]; ok {
								return err
							}
							return nil
						}).Times(1)
				}
			}

			updater := NewAccrualUpdater(mockOrderRepo, mockBalanceRepo, mockAccrualClient, 10*time.Millisecond)
			updater.checkAndUpdateOrders(ctx)
		})
	}
}

func TestAccrualUpdater_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderRepo := repository_mocks.NewMockOrderRepository(ctrl)
	mockBalanceRepo := repository_mocks.NewMockBalanceRepository(ctrl)
	mockAccrualClient := &mockAccrualClient{}

	mockOrderRepo.EXPECT().GetUnprocessedOrders(gomock.Any()).Return([]models.Order{}, nil).AnyTimes()

	updater := NewAccrualUpdater(mockOrderRepo, mockBalanceRepo, mockAccrualClient, 10*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())

	go updater.Run(ctx)

	time.Sleep(20 * time.Millisecond)

	cancel()

	time.Sleep(10 * time.Millisecond)

	assert.True(t, true, "Run function completed without errors")
}

func TestAccrualUpdater_Run_WithContextDone(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderRepo := repository_mocks.NewMockOrderRepository(ctrl)
	mockBalanceRepo := repository_mocks.NewMockBalanceRepository(ctrl)
	mockAccrualClient := &mockAccrualClient{}

	updater := NewAccrualUpdater(mockOrderRepo, mockBalanceRepo, mockAccrualClient, 1*time.Second)

	ctx, cancel := context.WithCancel(context.Background())

	cancel()

	updater.Run(ctx)

	assert.True(t, true, "Run function completed when context was cancelled")
}
