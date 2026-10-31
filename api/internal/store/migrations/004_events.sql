-- Registro de eventos e erros exibidos na tela Atividade.
CREATE TABLE IF NOT EXISTS events (
    id         BIGSERIAL PRIMARY KEY,
    level      TEXT NOT NULL DEFAULT 'info',
    source     TEXT NOT NULL DEFAULT 'app',
    message    TEXT NOT NULL,
    details    JSONB NOT NULL DEFAULT '{}'::jsonb,
    user_id    BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS events_created_idx ON events(created_at DESC);
CREATE INDEX IF NOT EXISTS events_level_idx ON events(level, created_at DESC);
