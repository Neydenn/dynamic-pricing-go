package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dynamic-pricing/config"
    "dynamic-pricing/internal/httpserver"
	"dynamic-pricing/internal/kafka"
	"dynamic-pricing/internal/order"
	"dynamic-pricing/internal/postgres"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "config.yaml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}

	db, err := postgres.NewPool(ctx, cfg.Order.DB)
	if err != nil {
		slog.Error("db", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	prod := kafka.NewProducer(cfg.Order.Kafka.Brokers, cfg.Order.Kafka.Topic)
	defer prod.Close()

	repo := order.NewRepository(db)
	svc := order.NewService(repo, prod)
	h := order.NewHandler(svc)

    srv := httpserver.New(cfg.Order.HTTPAddr, httpserver.CORS(h.Routes()))

	go func() {
		if err := srv.Start(); err != nil {
			slog.Error("http", "err", err)
			stop()
		}
	}()

	<-ctx.Done()

	shCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shCtx)
}
