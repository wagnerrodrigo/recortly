package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/recortly/recortly/api/internal/model"
)

func (s *Store) CreatePost(ctx context.Context, p model.Post) (model.Post, error) {
	err := s.Pool.QueryRow(ctx,
		`INSERT INTO posts (output_id, platform, title, description, scheduled_at, status)
		 VALUES ($1,$2,$3,$4,$5,$6)
		 RETURNING id, created_at`,
		p.OutputID, p.Platform, p.Title, p.Description, p.ScheduledAt, p.Status).
		Scan(&p.ID, &p.CreatedAt)
	return p, err
}

func (s *Store) GetPost(ctx context.Context, id int64) (model.Post, error) {
	var p model.Post
	err := s.Pool.QueryRow(ctx,
		`SELECT id, output_id, platform, title, description, scheduled_at, status,
		        external_id, external_url, error, created_at, published_at
		 FROM posts WHERE id=$1`, id).
		Scan(&p.ID, &p.OutputID, &p.Platform, &p.Title, &p.Description, &p.ScheduledAt,
			&p.Status, &p.ExternalID, &p.ExternalURL, &p.Error, &p.CreatedAt, &p.PublishedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return p, ErrNotFound
	}
	return p, err
}

func (s *Store) ListPosts(ctx context.Context, userID int64) ([]model.Post, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT p.id, p.output_id, p.platform, p.title, p.description, p.scheduled_at, p.status,
		        p.external_id, p.external_url, p.error, p.created_at, p.published_at
		 FROM posts p
		 JOIN render_outputs ro ON ro.id = p.output_id
		 JOIN renders r ON r.id = ro.render_id
		 JOIN clips c ON c.id = r.clip_id
		 JOIN sources s ON s.id = c.source_id
		 WHERE s.user_id=$1
		 ORDER BY p.created_at DESC LIMIT 100`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Post{}
	for rows.Next() {
		var p model.Post
		if err := rows.Scan(&p.ID, &p.OutputID, &p.Platform, &p.Title, &p.Description,
			&p.ScheduledAt, &p.Status, &p.ExternalID, &p.ExternalURL, &p.Error,
			&p.CreatedAt, &p.PublishedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) SetPostStatus(ctx context.Context, id int64, status, errMsg, externalID, externalURL string) error {
	var publishedAt *time.Time
	if status == "published" {
		now := time.Now()
		publishedAt = &now
	}
	_, err := s.Pool.Exec(ctx,
		`UPDATE posts SET status=$1::post_status, error=$2, external_id=$3, external_url=$4,
		        published_at = COALESCE($5, published_at)
		 WHERE id=$6`, status, errMsg, externalID, externalURL, publishedAt, id)
	return err
}

func (s *Store) PostForWorker(ctx context.Context, id int64) (model.Post, model.Source, model.RenderOutput, error) {
	var p model.Post
	err := s.Pool.QueryRow(ctx,
		`SELECT id, output_id, platform, title, description, scheduled_at, status,
		        external_id, external_url, error, created_at, published_at
		 FROM posts WHERE id=$1`, id).
		Scan(&p.ID, &p.OutputID, &p.Platform, &p.Title, &p.Description, &p.ScheduledAt,
			&p.Status, &p.ExternalID, &p.ExternalURL, &p.Error, &p.CreatedAt, &p.PublishedAt)
	if err != nil {
		return p, model.Source{}, model.RenderOutput{}, err
	}

	var o model.RenderOutput
	if err := s.Pool.QueryRow(ctx,
		`SELECT id, render_id, variant_idx, headline, file_path, duration_ms
		 FROM render_outputs WHERE id=$1`, p.OutputID).
		Scan(&o.ID, &o.RenderID, &o.VariantIdx, &o.Headline, &o.FilePath, &o.DurationMS); err != nil {
		return p, model.Source{}, o, err
	}
	o.FilePath = absPath(o.FilePath)

	row := s.Pool.QueryRow(ctx,
		`SELECT `+aliasedSourceCols+` FROM sources s
		 JOIN clips c ON c.source_id = s.id
		 JOIN renders r ON r.clip_id = c.id
		 WHERE r.id=$1`, o.RenderID)
	src, err := scanSource(row)
	return p, src, o, err
}

func (s *Store) UpsertAccount(ctx context.Context, userID int64, platform, label, access, refresh string, expires *time.Time) error {
	_, err := s.Pool.Exec(ctx,
		`INSERT INTO platform_accounts (user_id, platform, account_label, access_token, refresh_token, token_expires_at)
		 VALUES ($1,$2,$3,$4,$5,$6)
		 ON CONFLICT (user_id, platform) DO UPDATE SET
		   account_label=EXCLUDED.account_label,
		   access_token=EXCLUDED.access_token, refresh_token=EXCLUDED.refresh_token,
		   token_expires_at=EXCLUDED.token_expires_at`,
		userID, platform, label, access, refresh, expires)
	return err
}

func (s *Store) Account(ctx context.Context, userID int64, platform string) (model.PlatformAccount, string, error) {
	var a model.PlatformAccount
	var access string
	err := s.Pool.QueryRow(ctx,
		`SELECT id, user_id, platform, account_label, token_expires_at, access_token
		 FROM platform_accounts WHERE user_id=$1 AND platform=$2`, userID, platform).
		Scan(&a.ID, &a.UserID, &a.Platform, &a.AccountLabel, &a.TokenExpiresAt, &access)
	if errors.Is(err, pgx.ErrNoRows) {
		return a, "", ErrNotFound
	}
	a.Connected = err == nil && access != ""
	return a, access, err
}

func (s *Store) ListAccounts(ctx context.Context, userID int64) ([]model.PlatformAccount, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT id, user_id, platform, account_label, token_expires_at, access_token <> ''
		 FROM platform_accounts WHERE user_id=$1 ORDER BY platform`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.PlatformAccount{}
	for rows.Next() {
		var a model.PlatformAccount
		if err := rows.Scan(&a.ID, &a.UserID, &a.Platform, &a.AccountLabel, &a.TokenExpiresAt, &a.Connected); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) Log(ctx context.Context, userID *int64, action string, details map[string]any) {
	_, _ = s.Pool.Exec(ctx,
		`INSERT INTO audit_log (user_id, action, details) VALUES ($1,$2,$3)`,
		userID, action, details)
}

var _ = time.Now
