// devdb arranca um PostgreSQL embebido para desenvolvimento local sem Docker:
//
//	go run ./cmd/devdb
//	DATABASE_URL=postgres://postgres:postgres@localhost:5433/postgres?sslmode=disable go run ./cmd/server
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
)

func main() {
	runtimeDir := filepath.Join(os.TempDir(), "ft-devdb")
	pg := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
		Port(5433).
		RuntimePath(runtimeDir).
		DataPath(filepath.Join(runtimeDir, "data")).
		StartTimeout(90 * time.Second))
	if err := pg.Start(); err != nil {
		log.Fatalf("arrancar postgres embebido: %v", err)
	}
	fmt.Println("Postgres de desenvolvimento em localhost:5433 (user/pass/db: postgres) — Ctrl+C para parar")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	if err := pg.Stop(); err != nil {
		log.Fatal(err)
	}
}
