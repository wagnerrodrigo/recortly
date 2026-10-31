package store

import (
	"context"
	"encoding/json"
	"time"
)

type Event struct {
	ID        int64          `json:"id"`
	Level     string         `json:"level"`
	Source    string         `json:"source"`
	Message   string         `json:"message"`
	Details   map[string]any `json:"details"`
	UserID    *int64         `json:"userId,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
}

// AddEvent grava um evento (log) para a tela Atividade.
func (s *Store) AddEvent(ctx context.Context, level, source, message string, details map[string]any, userID *int64) {
	if details == nil {
		details = map[string]any{}
	}
	raw, err := json.Marshal(details)
	if err != nil {
		raw = []byte("{}")
	}
	_, _ = s.Pool.Exec(ctx,
		`INSERT INTO events (level, source, message, details, user_id) VALUES ($1,$2,$3,$4,$5)`,
		level, source, message, raw, userID)
}

func (s *Store) ListEvents(ctx context.Context, level, source string, limit int) ([]Event, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	rows, err := s.Pool.Query(ctx,
		`SELECT id, level, source, message, details, user_id, created_at
		 FROM events
		 WHERE ($1 = '' OR level = $1) AND ($2 = '' OR source = $2)
		 ORDER BY created_at DESC LIMIT $3`, level, source, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Event{}
	for rows.Next() {
		var e Event
		var raw []byte
		if err := rows.Scan(&e.ID, &e.Level, &e.Source, &e.Message, &raw, &e.UserID, &e.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(raw, &e.Details)
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) ClearEvents(ctx context.Context, olderThanDays int) (int64, error) {
	var tag interface{ RowsAffected() int64 }
	var err error
	if olderThanDays <= 0 {
		tag, err = s.Pool.Exec(ctx, `DELETE FROM events`)
	} else {
		tag, err = s.Pool.Exec(ctx,
			`DELETE FROM events WHERE created_at < now() - ($1 || ' days')::interval`, olderThanDays)
	}
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
