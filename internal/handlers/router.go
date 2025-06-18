package handlers

import (
	"github.com/a2sh3r/gophermart/internal/middleware"
	"github.com/a2sh3r/gophermart/internal/service"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type Handler struct {
	userService service.UserService
	secretKey   string
}

func NewHandler(userService service.UserService, secretKey string) *Handler {
	return &Handler{
		userService: userService,
		secretKey:   secretKey,
	}
}

func NewRouter(handler *Handler, secretKey string) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.NewLoggingMiddleware())
	r.Use(middleware.NewGzipMiddleware())
	r.Use(middleware.NewHashMiddleware(secretKey))

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Invalid URL format", http.StatusNotFound)
	})

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", handler.Register)
		r.Post("/login", handler.Login)
	})

	return r
}
