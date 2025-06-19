package service

import (
	"context"
	"errors"
	"github.com/a2sh3r/gophermart/internal/accrual"
	"github.com/a2sh3r/gophermart/internal/apperrors"
	repoMocks "github.com/a2sh3r/gophermart/internal/mocks/repository_mocks"
	"github.com/a2sh3r/gophermart/internal/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type fakeAccrualClient struct {
	resp       *accrual.AccrualResponse
	err        error
	statusCode int
}

func (f *fakeAccrualClient) GetOrderStatus(_ context.Context, _ string) (*accrual.AccrualResponse, int, error) {
	return f.resp, f.statusCode, f.err
}

func TestOrderService_UploadOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	userID := int64(1)
	tests := []struct {
		name          string
		orderNumber   string
		ownerID       int64
		ownerErr      error
		accrualResp   *accrual.AccrualResponse
		accrualErr    error
		accrualStatus int
		saveOrderErr  error
		expectedErr   error
	}{
		{
			name:        "валидный номер, новый заказ",
			orderNumber: "79927398713",
			ownerID:     0,
			accrualResp: &accrual.AccrualResponse{Status: StatusNew, Accrual: nil},
		},
		{
			name:        "невалидный номер заказа",
			orderNumber: "12345",
			expectedErr: apperrors.ErrInvalidOrderNumber,
		},
		{
			name:        "заказ уже пользователя",
			orderNumber: "79927398713",
			ownerID:     userID,
			expectedErr: apperrors.ErrOrderExistsSameUser,
		},
		{
			name:        "заказ другого пользователя",
			orderNumber: "79927398713",
			ownerID:     99,
			expectedErr: apperrors.ErrOrderExistsOtherUser,
		},
		{
			name:        "ошибка получения владельца",
			orderNumber: "79927398713",
			ownerErr:    errors.New("db error"),
			expectedErr: errors.New("db error"),
		},
		{
			name:         "ошибка сохранения заказа",
			orderNumber:  "79927398713",
			ownerID:      0,
			saveOrderErr: errors.New("save error"),
			expectedErr:  errors.New("save error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repoMocks.NewMockOrderRepository(ctrl)
			balanceRepo := repoMocks.NewMockBalanceRepository(ctrl)
			client := &fakeAccrualClient{
				resp:       tt.accrualResp,
				err:        tt.accrualErr,
				statusCode: tt.accrualStatus,
			}
			service := NewOrderService(repo, balanceRepo, client)

			if !errors.Is(tt.expectedErr, apperrors.ErrInvalidOrderNumber) {
				repo.EXPECT().GetOrderOwner(ctx, tt.orderNumber).Return(tt.ownerID, tt.ownerErr)
			}

			if tt.expectedErr == nil || tt.saveOrderErr != nil {
				repo.EXPECT().SaveOrder(ctx, gomock.Any()).Return(tt.saveOrderErr)
			}

			if tt.accrualResp != nil && tt.accrualResp.Status == accrual.StatusProcessed && tt.accrualResp.Accrual != nil && *tt.accrualResp.Accrual > 0 {
				balanceRepo.EXPECT().IncreaseUserBalance(ctx, userID, *tt.accrualResp.Accrual).Return(nil)
			}

			err := service.UploadOrder(ctx, tt.orderNumber, userID)

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOrderService_GetUserOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repoMocks.NewMockOrderRepository(ctrl)
	balanceRepo := repoMocks.NewMockBalanceRepository(ctrl)
	client := &fakeAccrualClient{}
	service := NewOrderService(repo, balanceRepo, client)
	ctx := context.Background()
	userID := int64(1)

	expectedOrders := []models.Order{{Number: "123", UserID: userID}}
	repo.EXPECT().GetOrdersByUser(ctx, userID).Return(expectedOrders, nil)

	orders, err := service.GetUserOrders(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, expectedOrders, orders)
}
