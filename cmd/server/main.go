package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"kart/internal/config"
	"kart/internal/repo"
	"kart/internal/server"
	"kart/internal/service"
	"kart/internal/sqlc"
	"kart/internal/store"
)

func main() {
	cfg := config.Load()
	db, err := store.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	// sqlc querier
	q := sqlc.New(db.DB)
	// repositories
	pr := repo.NewProductRepo(q)
	cr := repo.NewCouponRepo(q)
	or := repo.NewOrderRepo(db.DB)
	// services
	ps := service.NewProductService(pr)
	osvc := service.NewOrderService(pr, cr, or)

	h := &server.Server{Cfg: cfg, Products: ps, Orders: osvc}
	r, err := server.NewRouter(cfg.APIKey, h)
	if err != nil {
		log.Fatalf("router init: %v", err)
	}

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("env=%s listening on %s", cfg.Env, cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
		_ = srv.Close()
	}
	_ = db.Close()
}
