package handlers

import (
	"encoding/json"
	"errors"
	"github.com/a2sh3r/gophermart/internal/apperrors"
	"github.com/a2sh3r/gophermart/internal/logger"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"time"

	"net/http"
)

type authRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Login == "" || req.Password == "" {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	err := h.userService.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserAlreadyExists) {
			http.Error(w, "user already exists", http.StatusConflict)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		logger.Log.Error("register failed", zap.Error(err))
		return
	}

	user, err := h.userService.GetUserByLogin(r.Context(), req.Login)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		logger.Log.Error("get user failed", zap.Error(err))
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(h.secretKey))
	if err != nil {
		http.Error(w, "could not create token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+tokenString)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(authResponse{Token: tokenString})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Login == "" || req.Password == "" {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	err := h.userService.Authenticate(r.Context(), req.Login, req.Password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	user, err := h.userService.GetUserByLogin(r.Context(), req.Login)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		logger.Log.Error("get user failed", zap.Error(err))
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(h.secretKey))
	if err != nil {
		http.Error(w, "could not create token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+tokenString)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(authResponse{Token: tokenString})
}
