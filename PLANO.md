# SmashCut: Plano do Projeto

Ferramenta interna (uso meu e de um amigo) para transformar vídeos longos em cortes verticais e publicá-los no YouTube Shorts, Instagram Reels e TikTok.

Imagens de referência: `img/` (layout do painel, wizard de templates, cards de preview, paleta).

## 1. Escopo

**Dentro do escopo (MVP)**
- Login para 2 usuários.
- Importar vídeo próprio (upload).
- Editor de corte: marcar início/fim, cortar em vários trechos.
- Formato vertical 9:16 com enquadramento (centro, blur de fundo ou recorte manual).
- Texto/headline sobre o vídeo, com estilo salvo (fonte, cor, contorno, posição).
- Templates reutilizáveis (wizard: Perfil → Zonas → Frases → Mídia → Estilo → Variações → Gerar).
- Variações em lote a partir de um corte (zoom, cor, velocidade leve, espelhamento, trilha).
- Fila de render com progresso.
- Publicação e agendamento em YouTube, Instagram e TikTok.
- Registro de créditos e licença por vídeo (ver seção 6).

**Fora do escopo (por enquanto)**
- Scraping/download de Instagram ou TikTok de terceiros.
- Planos, cobrança e cotas.
- Tradução por IA e repost automático.
- Legenda automática (fase 5, via API de Whisper).
- MCP e comunidade.

## 2. Stack

| Camada | Escolha | Motivo |
|---|---|---|
| Frontend | Nuxt 4 + Tailwind + Nuxt UI | Já conhecido, SSR e componentes prontos |
| Backend (API) | Go + chi | Já conhecido, simples e rápido |
| Fila e worker | Go + asynq (Redis) | Mesmo binário de API ou worker separado, retries nativos |
| Banco | PostgreSQL + sqlc (ou GORM) | Consultas tipadas |
| Migrações | goose ou golang-migrate | Versionamento do schema |
| Vídeo | FFmpeg chamado via `os/exec` | Corte, crop, texto, variações |
| Auth | Sessão com cookie HttpOnly (2 usuários) | Simples, sem dependência externa |
| Storage | Disco local (`storage/`) → S3/R2 depois | Simples agora, migrável |
| Publicação | YouTube Data API (SDK Go oficial), Instagram Graph API, TikTok Content Posting API | APIs oficiais |
| Infra local | Docker Compose (Postgres + Redis) | Sobe com um comando |

Estrutura de pastas prevista:

```
smashcut/
├─ web/                 # Nuxt 4 (painel)
├─ api/                 # Go: HTTP (chi) e handlers
│  ├─ cmd/api/          # entrypoint da API
│  ├─ cmd/worker/       # entrypoint do worker asynq
│  └─ internal/
│     ├─ auth/
│     ├─ video/         # receitas FFmpeg e engine de templates
│     ├─ jobs/          # tarefas asynq (render, publish, cleanup)
│     ├─ publishers/    # YouTube, Instagram, TikTok
│     └─ store/         # acesso ao Postgres (sqlc)
├─ storage/             # vídeos de entrada e saída (fora do git)
├─ docker-compose.yml
└─ PLANO.md
```

## 3. Modelo de dados (resumo)

- `users`: id, nome, email, senha (hash bcrypt/argon2).
- `sources`: vídeos originais (path, duração, origem, autor, licença, permissão).
- `clips`: trecho de um source (início, fim, enquadramento, legenda).
- `templates`: nome, JSON com zonas, frases, estilo e variações.
- `renders`: clip + template, status (queued/running/done/failed), progresso, arquivo de saída.
- `posts`: render + plataforma, status, agendado para, id externo.
- `platform_accounts`: plataforma, tokens OAuth criptografados, usuário.
- `audit_log`: remoções e ações sensíveis.

## 4. Fases

**Fase 0: Fundação (1–2 dias)**
- Monorepo, Docker Compose, login com 2 usuários, layout base com sidebar (inspirado em `img/`).

**Fase 1: Corte manual (3–5 dias)**
- Upload de vídeo, player com marcação de in/out, lista de cortes.
- Render simples: corte + crop 9:16 com FFmpeg, download do arquivo.
- Fila asynq com progresso.

**Fase 2: Estilo e templates (4–6 dias)**
- Overlay de texto com estilo salvo.
- Wizard de template e preview em card (como nas imagens).
- Geração de variações em lote.

**Fase 3: Publicação (5–8 dias)**
- OAuth das contas.
- Publicação imediata e agendada.
- Começar por YouTube (mais simples), depois Instagram, depois TikTok.

**Fase 4: Refino**
- Biblioteca, busca, histórico, limpeza de arquivos antigos.

**Fase 5: Legenda automática (opcional)**
- Transcrição via API de Whisper (Groq ou OpenAI) chamada pelo Go.

## 5. Publicação: limitações conhecidas

- **YouTube**: upload via Data API. Shorts são vídeos verticais até 3 min com `#Shorts` no título ou descrição. Há cota diária de unidades da API.
- **Instagram**: Reels via Graph API, exige conta Business ou Creator e vídeo acessível por URL pública (upload temporário no R2/S3).
- **TikTok**: Content Posting API. Apps não auditados só conseguem postar com visibilidade privada até passarem na auditoria.

## 6. Direitos autorais (obrigatório no sistema)

- Cada `source` tem: autor, link original, tipo de licença (próprio / permissão por escrito / uso com crédito) e observações.
- Sem origem e licença preenchidas, o render não é liberado para publicação.
- A legenda recebe o crédito automaticamente.
- Guardar a permissão (print ou e-mail) junto ao registro quando houver.
- Função de remoção rápida de um vídeo e das publicações ligadas a ele.

Lembrete: "transformar" o conteúdo não garante uso legal. Use com permissão do criador sempre que possível.

## 7. Skills do projeto

Em `.opencode/skills/`:

| Skill | Origem | Para quê |
|---|---|---|
| `golang-patterns` | SkillsMP (ECC) | Padrões idiomáticos de Go |
| `golang-testing` | SkillsMP (ECC) | Testes em Go |
| `nuxt4-patterns` | SkillsMP (ECC) | Nuxt 4: hidratação, useFetch, SSR |
| `vue-patterns` | SkillsMP (ECC) | Composition API e Pinia |
| `ffmpeg` | SkillsMP (OpenMontage) | Operações de FFmpeg |
| `video-editing` | SkillsMP (ECC) | Fluxo de edição de vídeo |
| `smashcut-templates` | Própria | Schema dos templates e engine de variações |
| `smashcut-publishing` | Própria | Regras de publicação em YouTube, Instagram e TikTok |
| `smashcut-copyright` | Própria | Origem, licença e crédito obrigatórios |

## 8. Status da implementação

Implementado:

- [x] Fase 0: monorepo, Docker Compose, login com sessão em cookie e layout com sidebar.
- [x] Fase 1: upload com `ffprobe`, editor de cortes, render 9:16 e fila asynq com progresso.
- [x] Fluxo contínuo por template em 7 passos (Perfil, Zonas, Frases, Mídia, Estilo, Variações, Gerar).
- [x] Presets de template (Twitter/X, Meme, Vídeo viral, em branco) e zonas empilhadas com peso.
- [x] Headline com quebra automática de linha e faixa de fundo; perfil com avatar circular.
- [x] Variações com seed por render, recorte, zoom, saturação, brilho, velocidade e espelhamento.
- [x] Geração de frases com IA local (LM Studio/Ollama, API compatível com OpenAI).
- [x] Origem e licença obrigatórias, crédito automático e remoção de fonte.
- [x] YouTube: OAuth + upload via Data API (precisa de credenciais no `.env`).
- [x] Estabilidade do render: proxy H.264 automático para VP9/AV1/4K, watchdog de 90s,
      lock de worker único, recuperação no boot e botão "Tentar de novo".
- [x] Exclusão de renders (individual e em lote) preservando o histórico de publicações.
- [ ] Instagram e TikTok: adaptadores prontos, faltam app aprovado e credenciais.
- [ ] Zonas de imagem.
- [ ] Legenda automática por transcrição.

## 9. Próximos passos

1. Revisar este plano.
2. Pedir: "Faça a Fase 0 seguindo o PLANO.md".
3. Antes da Fase 3, criar os apps no YouTube, Meta e TikTok e anotar as credenciais.
