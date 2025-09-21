package main

import (
	"log"
	"net/http"

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
	// Migrations are handled externally (docker-compose migrate service)

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
	r := server.NewRouter(h)

	log.Printf("env=%s listening on %s", cfg.Env, cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
