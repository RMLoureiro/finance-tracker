package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"finance-tracker/server/internal/api"
	"finance-tracker/server/internal/config"
	"finance-tracker/server/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	st, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	if err := st.Migrate(ctx); err != nil {
		log.Fatalf("migrações: %v", err)
	}
	if err := st.SeedCategories(ctx); err != nil {
		log.Fatalf("seed de categorias: %v", err)
	}

	srv := api.New(st, cfg.SecureCookies)
	log.Printf("API a escutar em :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}
