package handlers

import (
	"github.com/a2sh3r/gophermart/internal/middleware"
	"github.com/a2sh3r/gophermart/internal/service"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type Handler struct {
	userService    service.UserService
	orderService   service.OrderService
	balanceService service.BalanceService
	secretKey      string
}

func NewHandler(
	userService service.UserService,
	orderService service.OrderService,
	balanceService service.BalanceService,
	secretKey string,
) *Handler {
	return &Handler{
		userService:    userService,
		orderService:   orderService,
		balanceService: balanceService,
		secretKey:      secretKey,
	}
}

func NewRouter(handler *Handler, secretKey string) chi.Router {
	r := chi.NewRouter()

	limiter := middleware.NewUserRateLimiter(1000, 1000)

	r.Use(middleware.NewLoggingMiddleware())
	r.Use(middleware.NewGzipMiddleware())
	r.Use(middleware.NewHashMiddleware(secretKey))
	r.Use(middleware.RateLimitMiddleware(limiter))

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Invalid URL format", http.StatusNotFound)
	})

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", handler.Register)
		r.Post("/login", handler.Login)

		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTMiddleware(secretKey))

			r.Post("/orders", handler.UploadOrder)
			r.Get("/orders", handler.GetOrders)
			r.Get("/balance", handler.GetBalance)
			r.Post("/balance/withdraw", handler.Withdraw)
			r.Get("/withdrawals", handler.GetWithdrawals)
		})
	})

	return r
}
