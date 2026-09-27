package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"qoshida/backend/internal/app"
	"qoshida/backend/internal/config"
	"qoshida/backend/internal/console"
)

func main() {
	log := console.New()

	cfg, err := config.Load()
	if err != nil {
		log.Fail("config", err)
		os.Exit(1)
	}

	log.Banner("QOSHIDA BACKEND", cfg.AppEnv, ":"+cfg.Port, cfg.DBHost, cfg.DBPort, cfg.DBName)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	application, err := app.New(ctx, cfg, log)
	if err != nil {
		log.Fail("startup", err)
		os.Exit(1)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- application.Run()
	}()

	log.Ready("http://localhost:" + cfg.Port)

	select {
	case <-ctx.Done():
		log.Info("shutdown", "signal qabul qilindi")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err = application.Shutdown(shutdownCtx); err != nil {
			log.Fail("shutdown", err)
			os.Exit(1)
		}
		log.OK("shutdown", "toza to'xtatildi")
	case err = <-errCh:
		if err != nil {
			log.Fail("server", err)
			os.Exit(1)
		}
	}
}
