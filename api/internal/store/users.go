package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/smashcut/smashcut/api/internal/model"
)

var ErrNotFound = errors.New("não encontrado")

func (s *Store) CreateUser(ctx context.Context, name, email, hash string) (model.User, error) {
	var u model.User
	err := s.Pool.QueryRow(ctx,
		`INSERT INTO users (name, email, password_hash) VALUES ($1,$2,$3)
		 ON CONFLICT (email) DO UPDATE SET name = users.name
		 RETURNING id, name, email, password_hash, created_at`,
		name, email, hash).
		Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt)
	return u, err
}

func (s *Store) UserByEmail(ctx context.Context, email string) (model.User, error) {
	var u model.User
	err := s.Pool.QueryRow(ctx,
		`SELECT id, name, email, password_hash, created_at FROM users WHERE email = $1`, email).
		Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}

func (s *Store) CreateSession(ctx context.Context, tokenHash string, userID int64, ttl time.Duration) error {
	_, err := s.Pool.Exec(ctx,
		`INSERT INTO sessions (token_hash, user_id, expires_at) VALUES ($1,$2,$3)`,
		tokenHash, userID, time.Now().Add(ttl))
	return err
}

func (s *Store) UserBySession(ctx context.Context, tokenHash string) (model.User, error) {
	var u model.User
	err := s.Pool.QueryRow(ctx,
		`SELECT u.id, u.name, u.email, u.password_hash, u.created_at
		 FROM sessions s JOIN users u ON u.id = s.user_id
		 WHERE s.token_hash = $1 AND s.expires_at > now()`, tokenHash).
		Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}

func (s *Store) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := s.Pool.Exec(ctx, `DELETE FROM sessions WHERE token_hash = $1`, tokenHash)
	return err
}
