package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/smashcut/smashcut/api/internal/model"
)

func (s *Store) CreateRender(ctx context.Context, clipID int64, templateID *int64, variants int) (model.Render, error) {
	var r model.Render
	seed := time.Now().UnixNano()
	err := s.Pool.QueryRow(ctx,
		`INSERT INTO renders (clip_id, template_id, variant_count, seed) VALUES ($1,$2,$3,$4)
		 RETURNING id, clip_id, template_id, variant_count, seed, status, progress, error, created_at, finished_at`,
		clipID, templateID, variants, seed).
		Scan(&r.ID, &r.ClipID, &r.TemplateID, &r.VariantCount, &r.Seed, &r.Status,
			&r.Progress, &r.Error, &r.CreatedAt, &r.FinishedAt)
	return r, err
}

func (s *Store) GetRender(ctx context.Context, userID, id int64) (model.Render, error) {
	var r model.Render
	err := s.Pool.QueryRow(ctx,
		`SELECT r.id, r.clip_id, r.template_id, r.variant_count, r.seed, r.status, r.progress, r.error,
		        r.created_at, r.finished_at
		 FROM renders r
		 JOIN clips c ON c.id = r.clip_id
		 JOIN sources s ON s.id = c.source_id
		 WHERE r.id=$1 AND s.user_id=$2`, id, userID).
		Scan(&r.ID, &r.ClipID, &r.TemplateID, &r.VariantCount, &r.Seed, &r.Status, &r.Progress,
			&r.Error, &r.CreatedAt, &r.FinishedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return r, ErrNotFound
	}
	if err != nil {
		return r, err
	}
	outs, err := s.RenderOutputs(ctx, r.ID)
	if err != nil {
		return r, err
	}
	r.Outputs = outs
	return r, nil
}

// RenderForWorker lê o render sem filtro de usuário (usado no worker).
func (s *Store) RenderForWorker(ctx context.Context, id int64) (model.Render, error) {
	var r model.Render
	err := s.Pool.QueryRow(ctx,
		`SELECT id, clip_id, template_id, variant_count, seed, status, progress, error, created_at, finished_at
		 FROM renders WHERE id=$1`, id).
		Scan(&r.ID, &r.ClipID, &r.TemplateID, &r.VariantCount, &r.Seed, &r.Status, &r.Progress,
			&r.Error, &r.CreatedAt, &r.FinishedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return r, ErrNotFound
	}
	return r, err
}

func (s *Store) ListRenders(ctx context.Context, userID int64, limit int) ([]model.Render, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := s.Pool.Query(ctx,
		`SELECT r.id, r.clip_id, r.template_id, r.variant_count, r.seed, r.status, r.progress, r.error,
		        r.created_at, r.finished_at, c.source_id
		 FROM renders r
		 JOIN clips c ON c.id = r.clip_id
		 JOIN sources s ON s.id = c.source_id
		 WHERE s.user_id=$1
		 ORDER BY r.created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Render{}
	for rows.Next() {
		var r model.Render
		var srcID int64
		if err := rows.Scan(&r.ID, &r.ClipID, &r.TemplateID, &r.VariantCount, &r.Seed, &r.Status,
			&r.Progress, &r.Error, &r.CreatedAt, &r.FinishedAt, &srcID); err != nil {
			return nil, err
		}
		r.SourceID = &srcID
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) SetRenderStatus(ctx context.Context, id int64, status string, progress int) error {
	var finishedAt *time.Time
	if status == "done" || status == "failed" || status == "cancelled" {
		now := time.Now()
		finishedAt = &now
	}
	_, err := s.Pool.Exec(ctx,
		`UPDATE renders SET status=$1::render_status, progress=$2, finished_at=COALESCE($3, finished_at)
		 WHERE id=$4`, status, progress, finishedAt, id)
	return err
}

func (s *Store) SetRenderError(ctx context.Context, id int64, msg string) error {
	_, err := s.Pool.Exec(ctx,
		`UPDATE renders SET status='failed'::render_status, error=$1, finished_at=now() WHERE id=$2`, msg, id)
	return err
}

func (s *Store) AddRenderOutput(ctx context.Context, out model.RenderOutput) (model.RenderOutput, error) {
	err := s.Pool.QueryRow(ctx,
		`INSERT INTO render_outputs (render_id, variant_idx, headline, file_path, duration_ms)
		 VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		out.RenderID, out.VariantIdx, out.Headline, out.FilePath, out.DurationMS).
		Scan(&out.ID)
	return out, err
}

func (s *Store) RenderOutputs(ctx context.Context, renderID int64) ([]model.RenderOutput, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT id, render_id, variant_idx, headline, file_path, duration_ms
		 FROM render_outputs WHERE render_id=$1 ORDER BY variant_idx`, renderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.RenderOutput{}
	for rows.Next() {
		var o model.RenderOutput
		if err := rows.Scan(&o.ID, &o.RenderID, &o.VariantIdx, &o.Headline, &o.FilePath, &o.DurationMS); err != nil {
			return nil, err
		}
		o.FilePath = absPath(o.FilePath)
		out = append(out, o)
	}
	return out, rows.Err()
}

func (s *Store) OutputForUser(ctx context.Context, userID, outputID int64) (model.RenderOutput, error) {
	var o model.RenderOutput
	err := s.Pool.QueryRow(ctx,
		`SELECT ro.id, ro.render_id, ro.variant_idx, ro.headline, ro.file_path, ro.duration_ms
		 FROM render_outputs ro
		 JOIN renders r ON r.id = ro.render_id
		 JOIN clips c ON c.id = r.clip_id
		 JOIN sources s ON s.id = c.source_id
		 WHERE ro.id=$1 AND s.user_id=$2`, outputID, userID).
		Scan(&o.ID, &o.RenderID, &o.VariantIdx, &o.Headline, &o.FilePath, &o.DurationMS)
	if errors.Is(err, pgx.ErrNoRows) {
		return o, ErrNotFound
	}
	if err == nil {
		o.FilePath = absPath(o.FilePath)
	}
	return o, err
}

func (s *Store) SourceForClip(ctx context.Context, clipID int64) (model.Source, error) {
	row := s.Pool.QueryRow(ctx,
		`SELECT `+aliasedSourceCols+` FROM sources s
		 JOIN clips c ON c.source_id = s.id WHERE c.id=$1`, clipID)
	return scanSource(row)
}

func (s *Store) DeleteOldRenders(ctx context.Context, days int) (int64, error) {
	tag, err := s.Pool.Exec(ctx,
		`DELETE FROM renders WHERE status='failed' AND created_at < now() - ($1 || ' days')::interval`,
		days)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// MarkStaleRendersFailed encerra renders que ficaram marcados como "running"
// quando o worker foi reiniciado ou travou.
func (s *Store) MarkStaleRendersFailed(ctx context.Context, message string) (int64, error) {
	tag, err := s.Pool.Exec(ctx,
		`UPDATE renders SET status='failed'::render_status, error=$1, finished_at=now()
		 WHERE status='running'`, message)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// QueuedRenders devolve renders parados na fila (para reenfileirar no boot).
func (s *Store) QueuedRenders(ctx context.Context) ([]int64, error) {
	rows, err := s.Pool.Query(ctx, `SELECT id FROM renders WHERE status='queued' ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []int64{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
