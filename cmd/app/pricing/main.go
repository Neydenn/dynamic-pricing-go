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
	"dynamic-pricing/internal/postgres"
	"dynamic-pricing/internal/pricing"
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

	db, err := postgres.NewPool(ctx, cfg.Pricing.DB)
	if err != nil {
		slog.Error("db", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	priceProd := kafka.NewProducer(cfg.Pricing.Kafka.Brokers, cfg.Pricing.Kafka.PricingTopic)
	defer priceProd.Close()

	repo := pricing.NewRepository(db)
	engine := pricing.NewEngine(repo, priceProd)

	catalogCons := kafka.NewConsumer(cfg.Pricing.Kafka.Brokers, cfg.Pricing.Kafka.CatalogTopic, cfg.Pricing.Kafka.GroupID+"-catalog")
	defer catalogCons.Close()

	ordersCons := kafka.NewConsumer(cfg.Pricing.Kafka.Brokers, cfg.Pricing.Kafka.OrdersTopic, cfg.Pricing.Kafka.GroupID+"-orders")
	defer ordersCons.Close()

	go func() {
		for {
			msg, err := catalogCons.Read(ctx)
			if err != nil {
				slog.Error("catalog-cons", "err", err)
				// retry on transient errors
				continue
			}
			if err := engine.HandleCatalogEvent(msg.Value); err != nil {
				slog.Error("catalog-ev", "err", err)
			}
		}
	}()

	go func() {
		for {
			msg, err := ordersCons.Read(ctx)
			if err != nil {
				slog.Error("orders-cons", "err", err)
				continue
			}
			if _, err := engine.HandleOrderEvent(ctx, msg.Value); err != nil {
				slog.Error("orders-ev", "err", err)
			}
		}
	}()

	h := pricing.NewHandler(repo, engine)
	srv := httpserver.New(cfg.Pricing.HTTPAddr, httpserver.CORS(h.Routes()))

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
