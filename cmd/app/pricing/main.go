package main

import (
    "context"
    "log/slog"
    "os"
    "os/signal"
    "syscall"

    "dynamic-pricing/config"
    "dynamic-pricing/internal/bootstrap"
)

func main() {
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "config.yaml"
	}

    cfg, err := config.Load(cfgPath)
    if err != nil { slog.Error("config", "err", err); os.Exit(1) }
    if err := bootstrap.RunPricing(ctx, cfg); err != nil { slog.Error("pricing", "err", err); os.Exit(1) }
}
