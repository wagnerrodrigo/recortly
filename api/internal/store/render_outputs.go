package store

import "context"

func (s *Store) DeleteRenderOutputs(ctx context.Context, renderID int64) error {
	_, err := s.Pool.Exec(ctx, `DELETE FROM render_outputs WHERE render_id=$1`, renderID)
	return err
}
