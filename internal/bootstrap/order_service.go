package bootstrap

import (
	"context"
	"log/slog"
	"time"

	"dynamic-pricing/config"
	"dynamic-pricing/internal/api/order_api"
	"dynamic-pricing/internal/httpserver"
	"dynamic-pricing/internal/producer"
	"dynamic-pricing/internal/services/order"
	"dynamic-pricing/internal/storage/pg"
)

func RunOrder(ctx context.Context, cfg config.Root) error {
	db, err := pg.NewPool(ctx, cfg.Order.DB)
	if err != nil {
		return err
	}
	defer db.Close()

	prod := producer.New(cfg.Order.Kafka.Brokers, cfg.Order.Kafka.Topic)
	defer prod.Close()

	repo := pg.NewOrderRepository(db)
	svc := order.NewService(repo, prod)
	h := order_api.NewHandler(svc)

	srv := httpserver.New(cfg.Order.HTTPAddr, httpserver.CORS(h.Routes()))
	go func() {
		if err := srv.Start(); err != nil {
			slog.Error("http", "err", err)
		}
	}()
	<-ctx.Done()
	shCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shCtx)
	return nil
}
