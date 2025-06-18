package main

import (
	"context"
	"github.com/a2sh3r/gophermart/internal/app"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	newApp, err := app.NewApp()
	if err != nil {
		panic(err)
	}

	if err := newApp.Run(ctx); err != nil {
		panic(err)
	}

	<-ctx.Done()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := newApp.Shutdown(shutdownCtx); err != nil {
		panic(err)
	}
}
