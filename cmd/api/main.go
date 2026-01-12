// @title Paku Booking API
// @version 1.0
// @description API del servicio Paku Booking
// @BasePath /
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpapi "paku-booking/internal/booking/http"
	"paku-booking/internal/config"
	"paku-booking/internal/jobs"
	"paku-booking/internal/outbox"
	"paku-booking/internal/shared"
	"paku-booking/internal/storage/memory"
)

func main() {
	cfg := config.Load()

	// Shared deps
	repo := memory.NewRepo()
	clk := shared.RealClock{}
	logger := shared.NewStdLogger()

	// HTTP Router (usa el MISMO repo)
	r := httpapi.NewRouter(httpapi.RouterOptions{
		Env:     cfg.Env,
		HoldTTL: cfg.HoldTTL,

		Repo:   repo,
		Clock:  clk,
		Logger: logger,
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// ---- Jobs (comparten repo) ----
	// 1) Expirar holds vencidos
	expirer := &jobs.HoldExpirer{
		Repo:     repo,
		Interval: 30 * time.Second, // ajusta a gusto
		Limit:    500,
		Clock:    clk,
		Logger:   logger,
	}

	// 2) Outbox dispatcher + publisher
	// MVP: NopPublisher (no envía nada hacia afuera)
	publisher := outbox.PublisherFromConfig(outbox.PublisherConfig{
		Env: cfg.Env,

		// opcional: si quieres setearlo por config en vez de env vars
		// HTTPEndpoint: cfg.OutboxEndpoint,
		// HTTPAPIKey:   cfg.OutboxAPIKey,
		// HTTPTimeout:  8 * time.Second,

		// DevFail: false,
	}, logger)

	dispatcher := &outbox.Dispatcher{
		Repo:      repo,
		Publisher: publisher,
		Interval:  5 * time.Second,
		BatchSize: 100,
		Clock:     clk,
		Logger:    logger,
	}

	go expirer.Run(ctx)
	go dispatcher.Run(ctx)

	// ---- Start server ----
	go func() {
		log.Printf("booking-api listening on :%s (env=%s)", cfg.Port, cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("shutting down...")
	_ = srv.Shutdown(shutdownCtx)
	log.Printf("bye")
}
