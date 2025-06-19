package handlers

import (
	"encoding/json"
	"errors"
	"github.com/a2sh3r/gophermart/internal/apperrors"
	"github.com/a2sh3r/gophermart/internal/logger"
	"github.com/a2sh3r/gophermart/internal/middleware"
	"github.com/a2sh3r/gophermart/internal/utils"
	"go.uber.org/zap"
	"io"
	"net/http"
	"regexp"
	"strings"
)

func (h *Handler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	orderNum := strings.TrimSpace(string(body))
	if !regexp.MustCompile(`^\d+$`).MatchString(orderNum) {
		http.Error(w, "invalid order format", http.StatusUnprocessableEntity)
		return
	}

	if !utils.IsValidLuhn(orderNum) {
		http.Error(w, "invalid order number", http.StatusUnprocessableEntity)
		return
	}

	err = h.orderService.UploadOrder(r.Context(), orderNum, userID)

	switch {
	case err == nil:
		w.WriteHeader(http.StatusAccepted)
	case errors.Is(err, apperrors.ErrOrderExistsSameUser):
		w.WriteHeader(http.StatusOK)
	case errors.Is(err, apperrors.ErrOrderExistsOtherUser):
		http.Error(w, "order already uploaded by another user", http.StatusConflict)
	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	orders, err := h.orderService.GetUserOrders(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(orders); err != nil {
		logger.Log.Error("failed to encode orders json", zap.Error(err))
	}
}
