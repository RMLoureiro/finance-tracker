package store

import (
	"context"
	"embed"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Store struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("ligar à base de dados: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping à base de dados: %w", err)
	}
	return &Store{Pool: pool}, nil
}

func (s *Store) Close() { s.Pool.Close() }

// Migrate aplica, por ordem, as migrações em migrations/ que ainda não constam
// de schema_migrations.
func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.Pool.Exec(ctx,
		`CREATE TABLE IF NOT EXISTS schema_migrations (name TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())`)
	if err != nil {
		return err
	}

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)

	for _, name := range names {
		var exists bool
		if err := s.Pool.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE name = $1)`, name).Scan(&exists); err != nil {
			return err
		}
		if exists {
			continue
		}
		sql, err := migrationsFS.ReadFile("migrations/" + name)
		if err != nil {
			return err
		}
		tx, err := s.Pool.Begin(ctx)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("migração %s: %w", name, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (name) VALUES ($1)`, name); err != nil {
			tx.Rollback(ctx)
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
	}
	return nil
}

// SeedCategories cria as categorias iniciais se ainda não existir nenhuma.
func (s *Store) SeedCategories(ctx context.Context) error {
	var n int
	if err := s.Pool.QueryRow(ctx, `SELECT count(*) FROM categories`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	seed := []struct {
		name, color, icon string
	}{
		{"Água", "#4A8AC9", "💧"},
		{"Eletricidade", "#B07E1F", "💡"},
		{"Gás", "#C25E37", "🔥"},
		{"Gasolina", "#8A64B8", "⛽"},
		{"Comida dos animais", "#0F9488", "🐾"},
		{"Supermercado", "#C94F72", "🛒"},
	}
	for i, c := range seed {
		if _, err := s.Pool.Exec(ctx,
			`INSERT INTO categories (name, color, icon, position) VALUES ($1, $2, $3, $4)`,
			c.name, c.color, c.icon, i); err != nil {
			return err
		}
	}
	return nil
}
