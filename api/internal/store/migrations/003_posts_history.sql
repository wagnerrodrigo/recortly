-- Histórico de publicações sobrevive à exclusão do render.
ALTER TABLE posts ALTER COLUMN output_id DROP NOT NULL;

DO $$
DECLARE cname text;
BEGIN
    SELECT conname INTO cname
    FROM pg_constraint
    WHERE conrelid = 'posts'::regclass AND contype = 'f'
    LIMIT 1;
    IF cname IS NOT NULL THEN
        EXECUTE format('ALTER TABLE posts DROP CONSTRAINT %I', cname);
    END IF;
END $$;

ALTER TABLE posts
    ADD CONSTRAINT posts_output_id_fkey
    FOREIGN KEY (output_id) REFERENCES render_outputs(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS posts_platform_idx ON posts(platform, created_at DESC);
