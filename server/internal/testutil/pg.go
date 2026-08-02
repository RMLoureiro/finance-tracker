// Package testutil arranca um PostgreSQL embebido para os testes, para não
// depender de um servidor local nem de Docker.
package testutil

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"

	"finance-tracker/server/internal/store"
)

// StartStore arranca um Postgres embebido, aplica as migrações e devolve um
// Store pronto a usar. Tudo é limpo no fim do teste.
func StartStore(t *testing.T) *store.Store {
	t.Helper()

	port := freePort(t)
	// runtime em os.TempDir() (sistema de ficheiros Linux): o Postgres exige
	// permissões 0700 no data dir, que não funcionam em /mnt/c.
	runtimeDir := filepath.Join(os.TempDir(), fmt.Sprintf("ft-pg-%d", port))

	pg := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
		Port(uint32(port)).
		RuntimePath(runtimeDir).
		StartTimeout(90 * time.Second).
		Logger(io.Discard))
	if err := pg.Start(); err != nil {
		t.Fatalf("arrancar postgres embebido: %v", err)
	}
	t.Cleanup(func() {
		pg.Stop()
		os.RemoveAll(runtimeDir)
	})

	url := fmt.Sprintf("postgres://postgres:postgres@localhost:%d/postgres?sslmode=disable", port)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	st, err := store.New(ctx, url)
	if err != nil {
		t.Fatalf("ligar ao postgres de teste: %v", err)
	}
	t.Cleanup(st.Close)

	if err := st.Migrate(ctx); err != nil {
		t.Fatalf("migrações: %v", err)
	}
	return st
}

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("obter porta livre: %v", err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}
