package bootstrap

import (
	"context"
	"dynamic-pricing/internal/api/catalog_api"
	"log/slog"
	"time"

	"dynamic-pricing/config"
	"dynamic-pricing/internal/httpserver"
	"dynamic-pricing/internal/producer"
	"dynamic-pricing/internal/services/catalog"
	"dynamic-pricing/internal/storage/pg"
)

func RunCatalog(ctx context.Context, cfg config.Root) error {
	db, err := pg.NewPool(ctx, cfg.Catalog.DB)
	if err != nil {
		return err
	}
	defer db.Close()

	prod := producer.New(cfg.Catalog.Kafka.Brokers, cfg.Catalog.Kafka.Topic)
	defer prod.Close()

	repo := pg.NewCatalogRepository(db)
	svc := catalog.NewService(repo, prod)
	h := catalog_api.NewHandler(svc)

	srv := httpserver.New(cfg.Catalog.HTTPAddr, httpserver.CORS(h.Routes()))
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
