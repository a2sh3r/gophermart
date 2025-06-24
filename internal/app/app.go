package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/a2sh3r/gophermart/internal/accrual"
	"github.com/a2sh3r/gophermart/internal/database"
	"github.com/a2sh3r/gophermart/internal/handlers"
	"github.com/a2sh3r/gophermart/internal/repository"
	"github.com/a2sh3r/gophermart/internal/service"
	"go.uber.org/zap"
	"net/http"
	"time"

	"github.com/a2sh3r/gophermart/internal/config"
	"github.com/a2sh3r/gophermart/internal/logger"
)

type App struct {
	server        *http.Server
	db            *sql.DB
	orderRepo     repository.OrderRepository
	balanceRepo   repository.BalanceRepository
	accrualClient *accrual.Client
}

func NewApp() (*App, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if err := logger.Initialize("debug"); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	db, err := database.InitDB(cfg)
	if err != nil {
		logger.Log.Error("Database connection failed", zap.Error(err))
		return nil, err
	}

	accrualClient := accrual.NewClient(cfg.AccrualSystemAddress)

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)

	orderRepo := repository.NewOrderRepository(db)
	balanceRepo := repository.NewBalanceRepository(db)
	orderService := service.NewOrderService(orderRepo, balanceRepo, accrualClient)

	balanceService := service.NewBalanceService(balanceRepo)

	handler := handlers.NewHandler(userService, orderService, balanceService, cfg.SecretKey)

	r := handlers.NewRouter(handler, cfg.SecretKey)

	server := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: r,
	}

	return &App{
		server:        server,
		db:            db,
		orderRepo:     orderRepo,
		balanceRepo:   balanceRepo,
		accrualClient: accrualClient,
	}, nil
}

func (a *App) Run(parentCtx context.Context) error {
	updater := service.NewAccrualUpdater(a.orderRepo, a.balanceRepo, a.accrualClient, time.Second*5)
	go updater.Run(parentCtx)

	serverErrCh := make(chan error, 1)
	go func() {
		err := a.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
			return
		}
		serverErrCh <- nil
	}()

	select {
	case <-parentCtx.Done():
		return nil
	case err := <-serverErrCh:
		return err
	}
}

func (a *App) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	logger.Log.Info("shutting down server...")
	if err := a.server.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error("server shutdown failed", zap.Error(err))
		return err
	}

	logger.Log.Info("closing database connection...")
	if err := a.db.Close(); err != nil {
		logger.Log.Error("failed to close database", zap.Error(err))
		return err
	}

	return nil
}
