-- Recortly: schema inicial
-- Idempotente: pode rodar em todo boot da API.

DO $$ BEGIN CREATE TYPE license_type AS ENUM ('own', 'permission', 'credit'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE layout_type AS ENUM ('fill', 'blur'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE render_status AS ENUM ('queued', 'running', 'done', 'failed', 'cancelled'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE platform_type AS ENUM ('youtube', 'instagram', 'tiktok'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE post_status AS ENUM ('draft', 'scheduled', 'publishing', 'published', 'failed'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;

CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    name          TEXT NOT NULL,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sessions (
    token_hash TEXT PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS sessions_user_idx ON sessions(user_id);

CREATE TABLE IF NOT EXISTS sources (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title           TEXT NOT NULL,
    file_path       TEXT NOT NULL,
    original_name   TEXT NOT NULL,
    size_bytes      BIGINT NOT NULL DEFAULT 0,
    duration_ms     BIGINT NOT NULL DEFAULT 0,
    width           INT NOT NULL DEFAULT 0,
    height          INT NOT NULL DEFAULT 0,
    has_audio       BOOLEAN NOT NULL DEFAULT true,
    author_name     TEXT NOT NULL DEFAULT '',
    author_handle   TEXT NOT NULL DEFAULT '',
    original_url    TEXT NOT NULL DEFAULT '',
    license_type    license_type,
    permission_proof TEXT NOT NULL DEFAULT '',
    notes           TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS sources_user_idx ON sources(user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS clips (
    id         BIGSERIAL PRIMARY KEY,
    source_id  BIGINT NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
    name       TEXT NOT NULL DEFAULT '',
    start_ms   BIGINT NOT NULL,
    end_ms     BIGINT NOT NULL,
    layout     layout_type NOT NULL DEFAULT 'blur',
    zoom       DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    offset_x   DOUBLE PRECISION NOT NULL DEFAULT 0.5,
    offset_y   DOUBLE PRECISION NOT NULL DEFAULT 0.5,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (end_ms > start_ms)
);
CREATE INDEX IF NOT EXISTS clips_source_idx ON clips(source_id);

CREATE TABLE IF NOT EXISTS templates (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    spec       JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS renders (
    id         BIGSERIAL PRIMARY KEY,
    clip_id    BIGINT NOT NULL REFERENCES clips(id) ON DELETE CASCADE,
    template_id BIGINT REFERENCES templates(id) ON DELETE SET NULL,
    variant_count INT NOT NULL DEFAULT 1 CHECK (variant_count BETWEEN 1 AND 30),
    status     render_status NOT NULL DEFAULT 'queued',
    progress   INT NOT NULL DEFAULT 0,
    error      TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS renders_status_idx ON renders(status, created_at DESC);

CREATE TABLE IF NOT EXISTS render_outputs (
    id          BIGSERIAL PRIMARY KEY,
    render_id   BIGINT NOT NULL REFERENCES renders(id) ON DELETE CASCADE,
    variant_idx INT NOT NULL,
    headline    TEXT NOT NULL DEFAULT '',
    file_path   TEXT NOT NULL DEFAULT '',
    duration_ms BIGINT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS render_outputs_render_idx ON render_outputs(render_id);

CREATE TABLE IF NOT EXISTS platform_accounts (
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    platform      platform_type NOT NULL,
    account_label TEXT NOT NULL DEFAULT '',
    access_token  TEXT NOT NULL DEFAULT '',
    refresh_token TEXT NOT NULL DEFAULT '',
    token_expires_at TIMESTAMPTZ,
    meta          JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, platform)
);

CREATE TABLE IF NOT EXISTS posts (
    id            BIGSERIAL PRIMARY KEY,
    output_id     BIGINT NOT NULL REFERENCES render_outputs(id) ON DELETE CASCADE,
    platform      platform_type NOT NULL,
    title         TEXT NOT NULL DEFAULT '',
    description   TEXT NOT NULL DEFAULT '',
    scheduled_at  TIMESTAMPTZ,
    status        post_status NOT NULL DEFAULT 'draft',
    external_id   TEXT NOT NULL DEFAULT '',
    external_url  TEXT NOT NULL DEFAULT '',
    error         TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at  TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS posts_schedule_idx ON posts(status, scheduled_at);

CREATE TABLE IF NOT EXISTS audit_log (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT REFERENCES users(id) ON DELETE SET NULL,
    action     TEXT NOT NULL,
    details    JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
