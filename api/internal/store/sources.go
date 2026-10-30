package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/smashcut/smashcut/api/internal/model"
)

const sourceCols = `id, user_id, title, file_path, original_name, size_bytes, duration_ms,
	width, height, has_audio, author_name, author_handle, original_url, license_type,
	permission_proof, notes, created_at`

func scanSource(row pgx.Row) (model.Source, error) {
	var s model.Source
	err := row.Scan(&s.ID, &s.UserID, &s.Title, &s.FilePath, &s.OriginalName, &s.SizeBytes,
		&s.DurationMS, &s.Width, &s.Height, &s.HasAudio, &s.AuthorName, &s.AuthorHandle,
		&s.OriginalURL, &s.LicenseType, &s.PermissionProof, &s.Notes, &s.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return s, ErrNotFound
	}
	s.FilePath = absPath(s.FilePath)
	return s, err
}

func (s *Store) CreateSource(ctx context.Context, in model.Source) (model.Source, error) {
	row := s.Pool.QueryRow(ctx,
		`INSERT INTO sources (user_id, title, file_path, original_name, size_bytes, duration_ms,
			width, height, has_audio, author_name, author_handle, original_url, license_type,
			permission_proof, notes)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		 RETURNING `+sourceCols,
		in.UserID, in.Title, in.FilePath, in.OriginalName, in.SizeBytes, in.DurationMS,
		in.Width, in.Height, in.HasAudio, in.AuthorName, in.AuthorHandle, in.OriginalURL,
		in.LicenseType, in.PermissionProof, in.Notes)
	return scanSource(row)
}

func (s *Store) UpdateSourceMeta(ctx context.Context, userID, id int64, in model.Source) error {
	_, err := s.Pool.Exec(ctx,
		`UPDATE sources SET author_name=$1, author_handle=$2, original_url=$3, license_type=$4,
		 permission_proof=$5, notes=$6, title=$7
		 WHERE id=$8 AND user_id=$9`,
		in.AuthorName, in.AuthorHandle, in.OriginalURL, in.LicenseType,
		in.PermissionProof, in.Notes, in.Title, id, userID)
	return err
}

func (s *Store) ListSources(ctx context.Context, userID int64) ([]model.Source, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT `+sourceCols+` FROM sources WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Source
	for rows.Next() {
		src, err := scanSource(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, src)
	}
	return out, rows.Err()
}

func (s *Store) GetSource(ctx context.Context, userID, id int64) (model.Source, error) {
	row := s.Pool.QueryRow(ctx,
		`SELECT `+sourceCols+` FROM sources WHERE id=$1 AND user_id=$2`, id, userID)
	return scanSource(row)
}

func (s *Store) DeleteSource(ctx context.Context, userID, id int64) error {
	_, err := s.Pool.Exec(ctx, `DELETE FROM sources WHERE id=$1 AND user_id=$2`, id, userID)
	return err
}

func (s *Store) CreateClip(ctx context.Context, c model.Clip) (model.Clip, error) {
	var out model.Clip
	err := s.Pool.QueryRow(ctx,
		`INSERT INTO clips (source_id, name, start_ms, end_ms, layout, zoom, offset_x, offset_y)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		 RETURNING id, source_id, name, start_ms, end_ms, layout, zoom, offset_x, offset_y, created_at`,
		c.SourceID, c.Name, c.StartMS, c.EndMS, c.Layout, c.Zoom, c.OffsetX, c.OffsetY).
		Scan(&out.ID, &out.SourceID, &out.Name, &out.StartMS, &out.EndMS, &out.Layout,
			&out.Zoom, &out.OffsetX, &out.OffsetY, &out.CreatedAt)
	return out, err
}

func (s *Store) ListClips(ctx context.Context, sourceID int64) ([]model.Clip, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT id, source_id, name, start_ms, end_ms, layout, zoom, offset_x, offset_y, created_at
		 FROM clips WHERE source_id=$1 ORDER BY start_ms`, sourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Clip
	for rows.Next() {
		var c model.Clip
		if err := rows.Scan(&c.ID, &c.SourceID, &c.Name, &c.StartMS, &c.EndMS, &c.Layout,
			&c.Zoom, &c.OffsetX, &c.OffsetY, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) GetClip(ctx context.Context, userID, id int64) (model.Clip, error) {
	var c model.Clip
	err := s.Pool.QueryRow(ctx,
		`SELECT c.id, c.source_id, c.name, c.start_ms, c.end_ms, c.layout, c.zoom,
		        c.offset_x, c.offset_y, c.created_at
		 FROM clips c JOIN sources s ON s.id = c.source_id
		 WHERE c.id=$1 AND s.user_id=$2`, id, userID).
		Scan(&c.ID, &c.SourceID, &c.Name, &c.StartMS, &c.EndMS, &c.Layout, &c.Zoom,
			&c.OffsetX, &c.OffsetY, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return c, ErrNotFound
	}
	return c, err
}

func (s *Store) DeleteClip(ctx context.Context, userID, id int64) error {
	_, err := s.Pool.Exec(ctx,
		`DELETE FROM clips c USING sources s WHERE c.source_id = s.id AND c.id=$1 AND s.user_id=$2`,
		id, userID)
	return err
}
