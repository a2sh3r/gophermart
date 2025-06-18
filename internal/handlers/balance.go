package handlers

import (
	"encoding/json"
	"errors"
	"github.com/a2sh3r/gophermart/internal/apperrors"
	"github.com/a2sh3r/gophermart/internal/logger"
	"github.com/a2sh3r/gophermart/internal/middleware"
	"github.com/a2sh3r/gophermart/internal/models"
	"go.uber.org/zap"
	"net/http"
)

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	balance, err := h.balanceService.GetUserBalance(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		logger.Log.Error("failed to get user balance", zap.Error(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(balance)
}

func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.WithdrawalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	err := h.balanceService.Withdraw(r.Context(), userID, req)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusOK)
	case errors.Is(err, apperrors.ErrInvalidOrderNumber):
		http.Error(w, "invalid order number", http.StatusUnprocessableEntity)
	case errors.Is(err, apperrors.ErrInsufficientFunds):
		http.Error(w, "insufficient funds", http.StatusPaymentRequired)
	case errors.Is(err, apperrors.ErrInvalidWithdrawalSum):
		http.Error(w, "invalid withdrawal sum", http.StatusBadRequest)
	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
		logger.Log.Error("withdraw error", zap.Error(err))
	}
}

func (h *Handler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.balanceService.GetWithdrawals(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		logger.Log.Error("failed to get withdrawals", zap.Error(err))
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(withdrawals); err != nil {
		logger.Log.Error("failed to encode withdrawals json", zap.Error(err))
	}
}
