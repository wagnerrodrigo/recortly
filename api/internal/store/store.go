package store

import (
	"context"
	"embed"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed all:migrations
var migrationsFS embed.FS

type Store struct {
	Pool *pgxpool.Pool
}

func Open(ctx context.Context, url string) (*Store, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres indisponível: %w", err)
	}
	return &Store{Pool: pool}, nil
}

func (s *Store) Close() { s.Pool.Close() }

// Migrate aplica o schema inicial (idempotente).
func Migrate(ctx context.Context, url string) error {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return err
	}
	defer pool.Close()

	sql, err := migrationsFS.ReadFile("migrations/001_init.sql")
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, string(sql)); err != nil {
		return fmt.Errorf("migração falhou: %w", err)
	}
	return nil
}

// MustStorage cria os diretórios de armazenamento.
func MustStorage(dir string) error {
	for _, sub := range []string{"sources", "renders", "tmp", "thumbnails"} {
		if err := os.MkdirAll(dir+"/"+sub, 0o755); err != nil {
			return err
		}
	}
	return nil
}
