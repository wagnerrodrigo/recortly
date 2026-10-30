package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/smashcut/smashcut/api/internal/model"
)

// SourceForOutput devolve a origem de um render_output validando o dono.
func (s *Store) SourceForOutput(ctx context.Context, userID, outputID int64) (model.Source, error) {
	row := s.Pool.QueryRow(ctx,
		`SELECT `+aliasedSourceCols+` FROM sources s
		 JOIN clips c ON c.source_id = s.id
		 JOIN renders r ON r.clip_id = c.id
		 JOIN render_outputs ro ON ro.render_id = r.id
		 WHERE ro.id=$1 AND s.user_id=$2`, outputID, userID)
	src, err := scanSource(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return src, ErrNotFound
	}
	return src, err
}

// ClipForOutput devolve o clip de um render_output.
func (s *Store) ClipForOutput(ctx context.Context, outputID int64) (model.Clip, error) {
	var c model.Clip
	err := s.Pool.QueryRow(ctx,
		`SELECT c.id, c.source_id, c.name, c.start_ms, c.end_ms, c.layout, c.zoom,
		        c.offset_x, c.offset_y, c.created_at
		 FROM clips c
		 JOIN renders r ON r.clip_id = c.id
		 JOIN render_outputs ro ON ro.render_id = r.id
		 WHERE ro.id=$1`, outputID).
		Scan(&c.ID, &c.SourceID, &c.Name, &c.StartMS, &c.EndMS, &c.Layout, &c.Zoom,
			&c.OffsetX, &c.OffsetY, &c.CreatedAt)
	return c, err
}
