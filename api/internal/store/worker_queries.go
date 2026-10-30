package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/smashcut/smashcut/api/internal/model"
)

func (s *Store) GetClipForWorker(ctx context.Context, id int64) (model.Clip, error) {
	var c model.Clip
	err := s.Pool.QueryRow(ctx,
		`SELECT id, source_id, name, start_ms, end_ms, layout, zoom, offset_x, offset_y, created_at
		 FROM clips WHERE id=$1`, id).
		Scan(&c.ID, &c.SourceID, &c.Name, &c.StartMS, &c.EndMS, &c.Layout, &c.Zoom,
			&c.OffsetX, &c.OffsetY, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return c, ErrNotFound
	}
	return c, err
}

func (s *Store) TemplateForWorker(ctx context.Context, id int64) (model.Template, error) {
	var t model.Template
	var raw []byte
	err := s.Pool.QueryRow(ctx,
		`SELECT id, user_id, name, spec, created_at, updated_at FROM templates WHERE id=$1`, id).
		Scan(&t.ID, &t.UserID, &t.Name, &raw, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return t, ErrNotFound
	}
	if err != nil {
		return t, err
	}
	return t, json.Unmarshal(raw, &t.Spec)
}
