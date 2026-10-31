package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Addr              string
	DatabaseURL       string
	RedisAddr         string
	SessionTTLHours   int
	StorageDir        string
	PublicBaseURL     string
	RenderConcurrency int
	TokenKey          string

	// Credenciais iniciais criadas no boot, se não existirem.
	SeedUsers []SeedUser

	// IA local (LM Studio, Ollama ou OpenAI compatível)
	AIBaseURL   string
	AIModel     string
	AIMaxTokens int
	AIReasoning string

	// OAuth
	YouTubeClientID     string
	YouTubeClientSecret string
	YouTubeRedirect     string
}

type SeedUser struct {
	Name     string
	Email    string
	Password string
}

func Load() Config {
	c := Config{
		Addr:                env("RECORTLY_ADDR", ":8080"),
		DatabaseURL:         env("RECORTLY_DATABASE_URL", "postgres://recortly:recortly@localhost:5433/recortly?sslmode=disable"),
		RedisAddr:           env("RECORTLY_REDIS_ADDR", "localhost:6380"),
		SessionTTLHours:     envInt("RECORTLY_SESSION_TTL_HOURS", 720),
		StorageDir:          env("RECORTLY_STORAGE_DIR", "storage"),
		PublicBaseURL:       strings.TrimRight(env("RECORTLY_PUBLIC_URL", "http://localhost:8080"), "/"),
		RenderConcurrency:   envInt("RECORTLY_RENDER_CONCURRENCY", 1),
		TokenKey:            env("RECORTLY_TOKEN_KEY", "dev-insecure-key-change-me"),
		AIBaseURL:           env("RECORTLY_AI_BASE_URL", "http://localhost:1234/v1"),
		AIModel:             env("RECORTLY_AI_MODEL", "gpt-oss-20b"),
		AIMaxTokens:         envInt("RECORTLY_AI_MAX_TOKENS", 3000),
		AIReasoning:         env("RECORTLY_AI_REASONING_EFFORT", "low"),
		YouTubeClientID:     env("RECORTLY_YOUTUBE_CLIENT_ID", ""),
		YouTubeClientSecret: env("RECORTLY_YOUTUBE_CLIENT_SECRET", ""),
		YouTubeRedirect:     env("RECORTLY_YOUTUBE_REDIRECT", "http://localhost:3000/api/oauth/youtube/callback"),
	}

	// RECORTLY_USERS="email:senha:Nome|email2:senha2:Nome2"
	if raw := env("RECORTLY_USERS", "eu@recortly.local:recortly123:Eu|amigo@recortly.local:recortly123:Amigo"); raw != "" {
		for _, part := range strings.Split(raw, "|") {
			fields := strings.Split(part, ":")
			if len(fields) < 2 || fields[0] == "" || fields[1] == "" {
				continue
			}
			name := strings.Split(fields[0], "@")[0]
			if len(fields) > 2 && fields[2] != "" {
				name = fields[2]
			}
			c.SeedUsers = append(c.SeedUsers, SeedUser{Email: fields[0], Password: fields[1], Name: name})
		}
	}
	if len(c.SeedUsers) == 0 {
		c.SeedUsers = []SeedUser{{Email: "eu@recortly.local", Password: "recortly123", Name: "Eu"}}
	}
	return c
}

func (c Config) AbsPath(parts ...string) string {
	return filepath.Join(append([]string{c.StorageDir}, parts...)...)
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	v, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return def
	}
	return v
}

func (c Config) String() string {
	return fmt.Sprintf("addr=%s storage=%s redis=%s", c.Addr, c.StorageDir, c.RedisAddr)
}
