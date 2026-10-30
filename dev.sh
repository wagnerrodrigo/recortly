#!/usr/bin/env bash
# Sobe banco, Redis, API, worker e frontend de uma vez.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"

if [ -f .env ]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

export SMASHCUT_DATABASE_URL="${SMASHCUT_DATABASE_URL:-postgres://smashcut:smashcut@localhost:5433/smashcut?sslmode=disable}"
export SMASHCUT_REDIS_ADDR="${SMASHCUT_REDIS_ADDR:-localhost:6380}"
export SMASHCUT_STORAGE_DIR="${SMASHCUT_STORAGE_DIR:-$ROOT/storage}"
export NUXT_API_PROXY_TARGET="${NUXT_API_PROXY_TARGET:-http://localhost:8080}"

if ! command -v ffmpeg >/dev/null 2>&1; then
  echo "ffmpeg não encontrado. Instale com: brew install ffmpeg" >&2
  exit 1
fi

echo "→ Subindo Postgres e Redis..."
docker compose up -d

echo "→ Aguardando o banco..."
for _ in $(seq 1 30); do
  if docker compose exec -T postgres pg_isready -U smashcut -d smashcut >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

mkdir -p storage/{sources,renders,tmp,thumbnails}

if [ ! -d web/node_modules ]; then
  echo "→ Instalando dependências do frontend..."
  (cd web && npm install)
fi

PIDS=()
cleanup() {
  echo ""
  echo "→ Encerrando..."
  for pid in "${PIDS[@]:-}"; do
    kill "$pid" 2>/dev/null || true
  done
  wait 2>/dev/null || true
}
trap cleanup INT TERM EXIT

# Escolhe uma porta livre para o painel (outros projetos podem ocupar a 3000).
WEB_PORT="${WEB_PORT:-3000}"
if lsof -nP -iTCP:"$WEB_PORT" -sTCP:LISTEN >/dev/null 2>&1; then
  for candidate in $(seq 3000 3010); do
    if ! lsof -nP -iTCP:"$candidate" -sTCP:LISTEN >/dev/null 2>&1; then
      WEB_PORT="$candidate"
      break
    fi
  done
  echo "→ Porta 3000 ocupada, usando $WEB_PORT"
fi

echo "→ API em http://localhost:8080"
(cd api && go run ./cmd/api) &
PIDS+=($!)

echo "→ Worker"
(cd api && go run ./cmd/worker) &
PIDS+=($!)

echo "→ Frontend em http://localhost:$WEB_PORT"
(cd web && npm run dev -- --port "$WEB_PORT") &
PIDS+=($!)

echo ""
echo "Pronto. Login: eu@smashcut.local / smashcut123 (troque no .env)"
wait
