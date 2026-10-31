package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/recortly/recortly/api/internal/model"
	"github.com/recortly/recortly/api/internal/video/template"
)

func (s *Store) CreateTemplate(ctx context.Context, userID int64, name string, spec template.Spec) (model.Template, error) {
	raw, err := json.Marshal(spec)
	if err != nil {
		return model.Template{}, err
	}
	var t model.Template
	var specRaw []byte
	err = s.Pool.QueryRow(ctx,
		`INSERT INTO templates (user_id, name, spec) VALUES ($1,$2,$3)
		 RETURNING id, user_id, name, spec, created_at, updated_at`,
		userID, name, raw).
		Scan(&t.ID, &t.UserID, &t.Name, &specRaw, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return t, err
	}
	return t, json.Unmarshal(specRaw, &t.Spec)
}

func (s *Store) UpdateTemplate(ctx context.Context, userID, id int64, name string, spec template.Spec) error {
	raw, err := json.Marshal(spec)
	if err != nil {
		return err
	}
	_, err = s.Pool.Exec(ctx,
		`UPDATE templates SET name=$1, spec=$2, updated_at=now() WHERE id=$3 AND user_id=$4`,
		name, raw, id, userID)
	return err
}

func (s *Store) ListTemplates(ctx context.Context, userID int64) ([]model.Template, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT id, user_id, name, spec, created_at, updated_at
		 FROM templates WHERE user_id=$1 ORDER BY updated_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Template{}
	for rows.Next() {
		var t model.Template
		var raw []byte
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &raw, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &t.Spec); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) GetTemplate(ctx context.Context, userID, id int64) (model.Template, error) {
	var t model.Template
	var raw []byte
	err := s.Pool.QueryRow(ctx,
		`SELECT id, user_id, name, spec, created_at, updated_at
		 FROM templates WHERE id=$1 AND user_id=$2`, id, userID).
		Scan(&t.ID, &t.UserID, &t.Name, &raw, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return t, ErrNotFound
	}
	if err != nil {
		return t, err
	}
	return t, json.Unmarshal(raw, &t.Spec)
}

func (s *Store) DeleteTemplate(ctx context.Context, userID, id int64) error {
	_, err := s.Pool.Exec(ctx, `DELETE FROM templates WHERE id=$1 AND user_id=$2`, id, userID)
	return err
}
