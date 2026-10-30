package store

// aliasedSourceCols qualifica as colunas de sources com o alias "s",
// necessário quando há JOIN com outras tabelas que também possuem "id".
const aliasedSourceCols = `s.id, s.user_id, s.title, s.file_path, s.original_name, s.size_bytes,
	s.duration_ms, s.width, s.height, s.has_audio, s.author_name, s.author_handle,
	s.original_url, s.license_type, s.permission_proof, s.notes, s.created_at`
