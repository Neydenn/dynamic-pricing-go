package bootstrap

import (
    "context"
    "log/slog"
    "time"

    "dynamic-pricing/config"
    "dynamic-pricing/internal/api/pricing_api"
    "dynamic-pricing/internal/httpserver"
    "dynamic-pricing/internal/consumer"
    "dynamic-pricing/internal/producer"
    "dynamic-pricing/internal/kafkautil"
    pricing "dynamic-pricing/internal/services/pricing"
    "dynamic-pricing/internal/storage/pg"
)

// RunPricing starts consumers, pricing engine, and HTTP read API until ctx is done.
func RunPricing(ctx context.Context, cfg config.Root) error {
    db, err := pg.NewPool(ctx, cfg.Pricing.DB)
    if err != nil { return err }
    defer db.Close()

    // Ensure the pricing topic exists to avoid UnknownTopic errors on publish.
    if err := kafkautil.EnsureTopic(ctx, cfg.Pricing.Kafka.Brokers, cfg.Pricing.Kafka.PricingTopic, 1, 1); err != nil {
        slog.Error("kafka ensure topic", "topic", cfg.Pricing.Kafka.PricingTopic, "err", err)
    }

    bus := producer.New(cfg.Pricing.Kafka.Brokers, cfg.Pricing.Kafka.PricingTopic)
    defer bus.Close()

    repo := pg.NewPriceRepository(db)
    eng := pricing.NewEngine(repo, bus)

    catalogCons := consumer.New(cfg.Pricing.Kafka.Brokers, cfg.Pricing.Kafka.CatalogTopic, cfg.Pricing.Kafka.GroupID+"-catalog")
    defer catalogCons.Close()

    ordersCons := consumer.New(cfg.Pricing.Kafka.Brokers, cfg.Pricing.Kafka.OrdersTopic, cfg.Pricing.Kafka.GroupID+"-orders")
    defer ordersCons.Close()

    go func() {
        for {
            msg, err := catalogCons.Read(ctx)
            if err != nil { slog.Error("catalog-cons", "err", err); continue }
            if err := eng.HandleCatalogEvent(msg.Value); err != nil { slog.Error("catalog-ev", "err", err) }
        }
    }()

    go func() {
        for {
            msg, err := ordersCons.Read(ctx)
            if err != nil { slog.Error("orders-cons", "err", err); continue }
            if _, err := eng.HandleOrderEvent(ctx, msg.Value); err != nil { slog.Error("orders-ev", "err", err) }
        }
    }()

    h := pricing_api.NewHandler(repo, eng)
    srv := httpserver.New(cfg.Pricing.HTTPAddr, httpserver.CORS(h.Routes()))
    go func() {
        if err := srv.Start(); err != nil { slog.Error("http", "err", err) }
    }()

    <-ctx.Done()
    shCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    _ = srv.Shutdown(shCtx)
    return nil
}
