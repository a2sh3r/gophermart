package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/a2sh3r/gophermart/internal/logger"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

type AccrualStatus string

const (
	StatusRegistered AccrualStatus = "REGISTERED"
	StatusInvalid    AccrualStatus = "INVALID"
	StatusProcessing AccrualStatus = "PROCESSING"
	StatusProcessed  AccrualStatus = "PROCESSED"
)

type ClientInterface interface {
	GetOrderStatus(ctx context.Context, number string) (*AccrualResponse, int, error)
}

type AccrualResponse struct {
	Order   string        `json:"order"`
	Status  AccrualStatus `json:"status"`
	Accrual *float64      `json:"accrual,omitempty"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) GetOrderStatus(ctx context.Context, orderNumber string) (*AccrualResponse, int, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNumber), nil)
	if err != nil {
		return nil, 0, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Log.Error("failed to close clients body", zap.Error(err))
		}
	}(resp.Body)

	logger.Log.Info("status code ", zap.Any("status", resp.StatusCode))

	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusTooManyRequests {
		return nil, resp.StatusCode, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result AccrualResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, resp.StatusCode, err
	}

	return &result, resp.StatusCode, nil
}
