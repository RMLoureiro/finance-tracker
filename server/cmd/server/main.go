package main

import (
	"context"
	"log"
	"net/http"
	"strings"
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
	handler := srv.Handler()
	if cfg.WebDir != "" {
		static := api.StaticHandler(cfg.WebDir)
		apiHandler := handler
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				apiHandler.ServeHTTP(w, r)
				return
			}
			static.ServeHTTP(w, r)
		})
	}

	log.Printf("a escutar em :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		log.Fatal(err)
	}
}
