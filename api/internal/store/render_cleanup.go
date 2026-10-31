package store

import (
	"context"
	"fmt"
	"time"
)

// RenderFiles são os caminhos de saída de um render.
type RenderFiles struct {
	RenderID int64
	Paths    []string
}

// DeleteRenders remove renders do usuário e devolve os arquivos a apagar.
// Regras:
//   - renders em execução ou na fila são ignorados (protegidos);
//   - publicações concluídas são mantidas no histórico (output_id = NULL);
//   - rascunhos, agendadas e falhas ligadas ao render são removidos.
func (s *Store) DeleteRenders(ctx context.Context, userID int64, ids []int64) ([]RenderFiles, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rows, err := tx.Query(ctx,
		`SELECT r.id, ro.file_path
		 FROM renders r
		 JOIN clips c ON c.id = r.clip_id
		 JOIN sources s ON s.id = c.source_id
		 LEFT JOIN render_outputs ro ON ro.render_id = r.id
		 WHERE s.user_id=$1 AND r.id = ANY($2) AND r.status NOT IN ('queued','running')`,
		userID, ids)
	if err != nil {
		return nil, err
	}
	files := []RenderFiles{}
	index := map[int64]int{}
	valid := []int64{}
	for rows.Next() {
		var id int64
		var path *string
		if err := rows.Scan(&id, &path); err != nil {
			rows.Close()
			return nil, err
		}
		if _, ok := index[id]; !ok {
			index[id] = len(files)
			files = append(files, RenderFiles{RenderID: id})
			valid = append(valid, id)
		}
		if path != nil && *path != "" {
			files[index[id]].Paths = append(files[index[id]].Paths, *path)
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(valid) == 0 {
		return files, nil
	}

	// Remove publicações que não chegaram a ser publicadas.
	if _, err := tx.Exec(ctx,
		`DELETE FROM posts WHERE output_id IN (
			SELECT ro.id FROM render_outputs ro WHERE ro.render_id = ANY($1)
		) AND status <> 'published'`, valid); err != nil {
		return nil, fmt.Errorf("remover publicações pendentes: %w", err)
	}

	// Desliga o histórico das publicadas (o ON DELETE SET NULL cobre o resto).
	if _, err := tx.Exec(ctx,
		`UPDATE posts SET output_id = NULL
		 WHERE output_id IN (SELECT ro.id FROM render_outputs ro WHERE ro.render_id = ANY($1))`,
		valid); err != nil {
		return nil, fmt.Errorf("preservar histórico: %w", err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM renders WHERE id = ANY($1)`, valid); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return files, nil
}

// RenderIDsForCleanup lista renders apagáveis de um usuário.
func (s *Store) RenderIDsForCleanup(ctx context.Context, userID int64, statuses []string, olderThanDays int) ([]int64, error) {
	if len(statuses) == 0 {
		statuses = []string{"done"}
	}
	var before *time.Time
	if olderThanDays > 0 {
		t := time.Now().AddDate(0, 0, -olderThanDays)
		before = &t
	}
	rows, err := s.Pool.Query(ctx,
		`SELECT r.id
		 FROM renders r
		 JOIN clips c ON c.id = r.clip_id
		 JOIN sources s ON s.id = c.source_id
		 WHERE s.user_id=$1
		   AND r.status = ANY($2)
		   AND ($3::timestamptz IS NULL OR r.created_at < $3)`,
		userID, statuses, before)
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

// ExistingRenderIDs devolve todos os IDs de render existentes.
func (s *Store) ExistingRenderIDs(ctx context.Context) (map[int64]bool, error) {
	rows, err := s.Pool.Query(ctx, `SELECT id FROM renders`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64]bool{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}
