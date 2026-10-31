-- Seed própria de cada render: garante variações diferentes a cada geração.
ALTER TABLE renders ADD COLUMN IF NOT EXISTS seed BIGINT NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS renders_clip_idx ON renders(clip_id, created_at DESC);
