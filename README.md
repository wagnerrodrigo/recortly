# SmashCut

Ferramenta pessoal para transformar vídeos longos em cortes verticais (9:16) e publicar no YouTube Shorts, Instagram Reels e TikTok.

Stack: **Nuxt 4** (painel) + **Go** (API e worker) + **PostgreSQL** + **Redis/asynq** + **FFmpeg**.

## Requisitos

- Go 1.24+
- Node 20+
- Docker
- FFmpeg (`brew install ffmpeg`)

## Começar

```bash
cp .env.example .env   # ajuste usuários, chaves e credenciais
make setup             # sobe banco/redis e instala dependências
make dev               # sobe API (:8080), worker e frontend (:3000)
```

Abra http://localhost:3000 e entre com um dos usuários do `.env`
(padrão: `eu@smashcut.local` / `smashcut123`). Se a porta 3000 estiver ocupada,
o script usa a próxima livre (3001, 3002...) e mostra no terminal.

> As portas do Postgres (5433) e Redis (6380) foram escolhidas para não conflitar com outros projetos locais.

## Como usar

1. **Biblioteca**: suba um vídeo e preencha **origem e licença** (autor, link, tipo de licença).
2. **Editor**: escolha o vídeo, marque início/fim, escolha o enquadramento e salve o corte.
3. **Templates**: monte o layout (perfil, headline, estilo e variações) e salve.
4. **Renders**: gere as variações, acompanhe o progresso e baixe os arquivos.
5. **Publicações**: publique agora, agende ou salve como rascunho.
6. **Integrações**: conecte o YouTube (OAuth) para publicar de verdade.

## Estrutura

```
api/
  cmd/api/          # servidor HTTP (chi)
  cmd/worker/       # consumidor asynq (FFmpeg + publicação)
  internal/
    config/         # variáveis de ambiente
    httpapi/        # rotas e handlers
    jobs/           # tasks asynq
    model/          # tipos do domínio
    publishers/     # YouTube, Instagram, TikTok
    store/          # Postgres (pgx) + migrações embutidas
    video/          # montagem do comando FFmpeg e variações
web/
  app/pages/        # telas do painel
  app/composables/  # sessão e utilitários de API
  server/routes/api # proxy para a API Go
docker-compose.yml
PLANO.md            # escopo, fases e decisões
```

## Variáveis principais

| Variável | Padrão | Descrição |
|---|---|---|
| `SMASHCUT_DATABASE_URL` | `postgres://smashcut:smashcut@localhost:5433/smashcut` | Banco |
| `SMASHCUT_REDIS_ADDR` | `localhost:6380` | Redis da fila |
| `SMASHCUT_USERS` | 2 usuários de exemplo | `email:senha:Nome|email:senha:Nome` |
| `SMASHCUT_STORAGE_DIR` | `./storage` | Vídeos de entrada e saída |
| `SMASHCUT_RENDER_CONCURRENCY` | `1` | Renders simultâneos |
| `SMASHCUT_YOUTUBE_CLIENT_ID` | vazio | OAuth do YouTube |
| `SMASHCUT_YOUTUBE_CLIENT_SECRET` | vazio | OAuth do YouTube |
| `SMASHCUT_YOUTUBE_REDIRECT` | `http://localhost:3000/api/oauth/youtube/callback` | Redirect |
| `SMASHCUT_TOKEN_KEY` | valor de dev | Assina o `state` do OAuth |

## Direitos autorais

O app bloqueia a publicação sem origem e licença preenchidas e adiciona o crédito
automaticamente na descrição. Publique apenas conteúdo próprio ou com permissão do criador.

## Limitações atuais

- **Instagram** e **TikTok** ainda não publicam: os adaptadores existem e retornam uma
  mensagem clara. É necessário app aprovado nas plataformas.
- **YouTube** publica via Data API quando as credenciais OAuth estão configuradas.
- O preview dos templates é aproximado; o resultado final vem do FFmpeg.
- Não há download de conteúdo de terceiros (por decisão de escopo e de direitos).

## Testes manuais do fluxo

```bash
# API
curl localhost:8080/api/health

# worker + render
curl -b cookies.txt -X POST localhost:8080/api/renders \
  -H 'Content-Type: application/json' \
  -d '{"clipId":1,"templateId":1,"variantCount":3}'
```

O plano completo, com fases e decisões de produto, está em [`PLANO.md`](./PLANO.md).

## Manutenção

```bash
make db-reset     # apaga o banco e recria (perde os dados)
rm -rf storage/*  # limpa vídeos de entrada, renders e temporários
```

Os renders ficam em `storage/renders/<id>/`; o job diário de limpeza remove
temporários com mais de 7 dias.
