package accrual

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient_GetOrderStatus(t *testing.T) {
	type want struct {
		resp       *AccrualResponse
		statusCode int
		err        bool
	}
	tests := []struct {
		name           string
		serverResponse string
		serverStatus   int
		want           want
	}{
		{
			name:           "успешный ответ с начислением",
			serverResponse: `{"order":"123","status":"PROCESSED","accrual":100.5}`,
			serverStatus:   http.StatusOK,
			want: want{
				resp:       &AccrualResponse{Order: "123", Status: StatusProcessed, Accrual: float64Ptr(100.5)},
				statusCode: http.StatusOK,
				err:        false,
			},
		},
		{
			name:           "успешный ответ без начисления",
			serverResponse: `{"order":"123","status":"PROCESSING"}`,
			serverStatus:   http.StatusOK,
			want: want{
				resp:       &AccrualResponse{Order: "123", Status: StatusProcessing, Accrual: nil},
				statusCode: http.StatusOK,
				err:        false,
			},
		},
		{
			name:           "нет данных",
			serverResponse: "",
			serverStatus:   http.StatusNoContent,
			want: want{
				resp:       nil,
				statusCode: http.StatusNoContent,
				err:        false,
			},
		},
		{
			name:           "ошибка сервера",
			serverResponse: "",
			serverStatus:   http.StatusInternalServerError,
			want: want{
				resp:       nil,
				statusCode: http.StatusInternalServerError,
				err:        true,
			},
		},
		{
			name:           "невалидный json",
			serverResponse: `{"order":123,"status":}`,
			serverStatus:   http.StatusOK,
			want: want{
				resp:       nil,
				statusCode: http.StatusOK,
				err:        true,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
				_, _ = w.Write([]byte(tt.serverResponse))
			}))
			defer srv.Close()

			client := NewClient(srv.URL)
			client.httpClient.Timeout = 2 * time.Second

			resp, status, err := client.GetOrderStatus(context.Background(), "123")
			if tt.want.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want.statusCode, status)
			if tt.want.resp != nil {
				assert.Equal(t, tt.want.resp.Order, resp.Order)
				assert.Equal(t, tt.want.resp.Status, resp.Status)
				if tt.want.resp.Accrual != nil {
					assert.NotNil(t, resp.Accrual)
					assert.InDelta(t, *tt.want.resp.Accrual, *resp.Accrual, 0.001)
				} else {
					assert.Nil(t, resp.Accrual)
				}
			} else {
				assert.Nil(t, resp)
			}
		})
	}
}

func float64Ptr(f float64) *float64 {
	return &f
}
