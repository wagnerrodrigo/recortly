package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Addr            string
	DatabaseURL     string
	RedisAddr       string
	SessionTTLHours int
	StorageDir      string
	PublicBaseURL   string
	RenderConcurrency int
	TokenKey        string

	// Credenciais iniciais criadas no boot, se não existirem.
	SeedUsers []SeedUser

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
		Addr:              env("SMASHCUT_ADDR", ":8080"),
		DatabaseURL:       env("SMASHCUT_DATABASE_URL", "postgres://smashcut:smashcut@localhost:5433/smashcut?sslmode=disable"),
		RedisAddr:         env("SMASHCUT_REDIS_ADDR", "localhost:6380"),
		SessionTTLHours:   envInt("SMASHCUT_SESSION_TTL_HOURS", 720),
		StorageDir:        env("SMASHCUT_STORAGE_DIR", "storage"),
		PublicBaseURL:     strings.TrimRight(env("SMASHCUT_PUBLIC_URL", "http://localhost:8080"), "/"),
		RenderConcurrency: envInt("SMASHCUT_RENDER_CONCURRENCY", 1),
		TokenKey:          env("SMASHCUT_TOKEN_KEY", "dev-insecure-key-change-me"),
		YouTubeClientID:   env("SMASHCUT_YOUTUBE_CLIENT_ID", ""),
		YouTubeClientSecret: env("SMASHCUT_YOUTUBE_CLIENT_SECRET", ""),
		YouTubeRedirect:   env("SMASHCUT_YOUTUBE_REDIRECT", "http://localhost:3000/api/oauth/youtube/callback"),
	}

	// SMASHCUT_USERS="email:senha:Nome|email2:senha2:Nome2"
	if raw := env("SMASHCUT_USERS", "eu@smashcut.local:smashcut123:Eu|amigo@smashcut.local:smashcut123:Amigo"); raw != "" {
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
		c.SeedUsers = []SeedUser{{Email: "eu@smashcut.local", Password: "smashcut123", Name: "Eu"}}
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
