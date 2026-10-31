# Recortly

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
(padrão: `eu@recortly.local` / `recortly123`). Se a porta 3000 estiver ocupada,
o script usa a próxima livre (3001, 3002...) e mostra no terminal.

> As portas do Postgres (5433) e Redis (6380) foram escolhidas para não conflitar com outros projetos locais.

## Como usar

O caminho principal é o **template**, com 7 passos:

**Templates → Perfil → Zonas → Frases → Mídia → Estilo → Variações → Gerar**

1. **Biblioteca**: suba o vídeo e preencha **origem e licença** (autor, link, tipo de licença).
2. **Templates**: escolha um modelo (Twitter/X, Meme, Vídeo viral ou em branco) e siga os passos.
   - **Zonas**: empilhe **vídeo** (principal ou outro da biblioteca), **imagem** (upload)
     e espaço vazio, com peso e cantos arredondados.
   - **Frases**: escreva ou gere com IA local.
   - **Mídia**: escolha o vídeo e marque o corte ali mesmo.
   - **Estilo**: fonte, cor, contorno e faixa. Cada etapa tem "Restaurar padrão".
   - **Variações**: quantidade, faixas de zoom/cor/recorte e espelhamento; cada geração usa uma seed nova.
   - **Perfil**: inclui o **avatar/logo** — a imagem redonda do perfil. Sem imagem, aparecem as iniciais do nome.
   - **Gerar**: dispara o render e mostra o progresso e os arquivos na própria tela.
3. **Editor** (opcional): corte rápido para quem prefere a timeline simplificada.
4. **Renders**: veja todos os renders, baixe ou vá para publicar.
5. **Publicações**: publique agora, agende ou salve como rascunho.
6. **Integrações**: conecte o YouTube (OAuth) para publicar de verdade.
7. **Atividade**: logs e erros de render, publicação, vídeo e API. É onde você confere
   por que algo travou, com filtros por nível, atualização automática e botão para
   **destravar** renders presos.

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
| `RECORTLY_DATABASE_URL` | `postgres://recortly:recortly@localhost:5433/recortly` | Banco |
| `RECORTLY_REDIS_ADDR` | `localhost:6380` | Redis da fila |
| `RECORTLY_USERS` | 2 usuários de exemplo | `email:senha:Nome|email:senha:Nome` |
| `RECORTLY_STORAGE_DIR` | `./storage` | Vídeos de entrada e saída |
| `RECORTLY_RENDER_CONCURRENCY` | `1` | Renders simultâneos |
| `RECORTLY_YOUTUBE_CLIENT_ID` | vazio | OAuth do YouTube |
| `RECORTLY_YOUTUBE_CLIENT_SECRET` | vazio | OAuth do YouTube |
| `RECORTLY_YOUTUBE_REDIRECT` | `http://localhost:3000/api/oauth/youtube/callback` | Redirect |
| `RECORTLY_TOKEN_KEY` | valor de dev | Assina o `state` do OAuth |
| `RECORTLY_AI_BASE_URL` | `http://localhost:1234/v1` | API de IA compatível com OpenAI |
| `RECORTLY_AI_MODEL` | `gpt-oss-20b` | Modelo usado nas frases |
| `RECORTLY_AI_MAX_TOKENS` | `3000` | Limite de resposta (modelos de raciocínio precisam de folga) |
| `RECORTLY_AI_REASONING_EFFORT` | `low` | Esforço de raciocínio (low/medium/high) |

## IA local (frases)

O passo **Frases** pode gerar headlines com o modelo rodando na sua máquina:

```bash
# LM Studio: ative o servidor local (Developer → Start Server) e use o id exato do modelo
RECORTLY_AI_BASE_URL=http://localhost:1234/v1
RECORTLY_AI_MODEL=openai/gpt-oss-20b   # confira o id em /v1/models

# Ollama
RECORTLY_AI_BASE_URL=http://localhost:11434/v1
RECORTLY_AI_MODEL=gpt-oss:20b
```

Sem o servidor ligado, a tela mostra a mensagem de erro e o restante do fluxo continua funcionando.
Você pode conferir o status em `GET /api/ai/status`.

Se a resposta vier vazia com `finish_reason: length`, o modelo gastou os tokens em raciocínio.
Aumente `RECORTLY_AI_MAX_TOKENS` (ex.: 4000) ou reduza `RECORTLY_AI_REASONING_EFFORT`.

## Direitos autorais

O app bloqueia a publicação sem origem e licença preenchidas e adiciona o crédito
automaticamente na descrição. Publique apenas conteúdo próprio ou com permissão do criador.

## Limitações atuais

- **Instagram** e **TikTok** ainda não publicam: os adaptadores existem e retornam uma
  mensagem clara. É necessário app aprovado nas plataformas.
- **YouTube** publica via Data API quando as credenciais OAuth estão configuradas.
- O preview dos templates é aproximado; o resultado final vem do FFmpeg. Ele tem
play/pause e barra de tempo para conferir o corte antes de renderizar.
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

## Nomes dos downloads

Cada arquivo baixado recebe o nome da **frase** usada naquela variação:

- `Olha so o que rolou nesse servidor.mp4` quando há uma variação;
- `01 - Primeira frase.mp4`, `02 - Segunda frase.mp4` quando há várias;
- **Baixar tudo (.zip)** baixa todas as variações de um render em um ZIP com os mesmos nomes.

Acentos são removidos e caracteres inválidos para nome de arquivo são substituídos.

## Atividade (logs e erros)

A tela **Atividade** mostra eventos gravados pelo worker e pela API:

- falhas de render (FFmpeg, proxy, template, arquivo);
- falhas e bloqueios de publicação;
- erros internos da API (5xx);
- renders concluídos (informativo).

Tem filtro por nível (erro/aviso/info), atualização automática a cada 5s, detalhes em JSON
e limpeza por idade. Se houver render em `running` por mais de 10 minutos, aparece um aviso
com o botão **Destravar**.

## Apagar renders

Na tela **Renders**:

- **Apagar** (em cada render): remove os arquivos e o registro. Fica desabilitado enquanto o
  render está na fila ou em execução.
- **Apagar antigos**: modal para remover em lote renders com falha/cancelados e/ou concluídos
  com mais de N dias.

Regras:

- Publicações **já concluídas continuam no histórico** de Publicações, com título, plataforma,
  data e link. O `output_id` fica nulo, mas o registro permanece.
- Rascunhos, agendamentos e publicações com falha ligados ao render são removidos.
- **Nada é apagado no YouTube, Instagram ou TikTok.** O app não chama API de exclusão.
- Pastas órfãs em `storage/renders/` (sem registro no banco) são limpas junto com o purge.

## Problemas comuns

**Render fica em "running" e não termina**

O worker tem proteção em três camadas:

1. **Proxy automático**: vídeos VP9/AV1/4K são convertidos uma vez para H.264 1080p em
   `storage/proxies/<source-id>.mp4`. O primeiro render demora mais; os próximos reaproveitam.
2. **Watchdog**: se o FFmpeg ficar 90s sem progresso, o processo é encerrado e o render
   falha com mensagem clara em vez de travar a fila para sempre.
3. **Recuperação no boot**: ao iniciar, o worker marca renders presos em `running` como
   falha e reenfileira os que estavam na fila.

Na tela de Renders, renders com falha têm o botão **"Tentar de novo"**.
Também existe `POST /api/renders/recover` para destravar manualmente.

**Só pode existir um worker**

O worker usa um lock em `storage/.worker.lock`. Se você subir dois (`go run` + `dev.sh`),
o segundo sai com a mensagem `já existe um worker rodando` — era isso que causava renders
duplicados e travados.

**Parar tudo**

```bash
./dev.sh stop      # encerra API, worker, frontend e containers
```

## Manutenção

```bash
make db-reset     # apaga o banco e recria (perde os dados)
rm -rf storage/*  # limpa vídeos de entrada, renders e temporários
```

Os renders ficam em `storage/renders/<id>/`; o job diário de limpeza remove
temporários com mais de 7 dias.
