package graceful

import (
	"context"
	"markdown-enricher/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Application interface {
	RunAndWait()
}

type Graceful struct {
	StartAction    func() error
	DeferAction    func(ctx context.Context) error
	ShutdownAction func(ctx context.Context) error
}

func (graceful *Graceful) RunAndWait() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := graceful.StartAction(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("listen: %s\n", err)
		}
	}()
	logger.Info("Server Started")

	<-done
	logger.Info("Server Stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		if err := graceful.DeferAction(ctx); err != nil {
			logger.Errorf("Server  Failed:%+v", err)
		}
		cancel()
	}()

	if err := graceful.ShutdownAction(ctx); err != nil {
		logger.Errorf("Server Shutdown Failed:%+v", err)
	}
	logger.Info("Server Exited Properly")
}
