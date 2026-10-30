---
name: smashcut-publishing
description: Use when implementing OAuth connections, scheduling, or publishing SmashCut clips to YouTube Shorts, Instagram Reels, or TikTok through official APIs.
---

# SmashCut: Publicação

Use **somente APIs oficiais**. Não automatize o app/site das redes.

## Interface comum (Go)

Cada plataforma é um adaptador em `api/internal/publishers/` implementando:

```go
type Publisher interface {
    Platform() string
    Connect(ctx context.Context, userID int64) (authURL string, err error)
    HandleCallback(ctx context.Context, code string, userID int64) error
    Publish(ctx context.Context, in PublishInput) (PublishResult, error)
}

type PublishInput struct {
    VideoPath   string
    Title       string
    Description string    // deve incluir crédito de origem (ver smashcut-copyright)
    ScheduledAt *time.Time
    Visibility  string    // public | unlisted | private
}

type PublishResult struct {
    ExternalID string
    URL        string
}
```

## Tokens

- Guardar em `platform_accounts` com criptografia AES-256-GCM. Chave em `.env` (`TOKEN_ENCRYPTION_KEY`).
- Renovar refresh token antes de expirar. Se falhar, marcar a conta como "reconectar".
- Nunca logar tokens.

## YouTube (começar aqui)

- API: YouTube Data API v3, `videos.insert`.
- Escopos: `youtube.upload`, `youtube.readonly`.
- Shorts: vídeo vertical até 3 min. Incluir `#Shorts` no título ou descrição.
- Cota diária: upload custa bastante unidade. Mostrar aviso se estiver perto do limite.
- Agendamento: `status.publishAt` com `privacyStatus: "private"`.

## Instagram (Reels)

- API: Instagram Graph API (conta Business ou Creator ligada a uma Página do Facebook).
- Fluxo: criar container (`media_type=REELS`, `video_url`) → aguardar status `FINISHED` → `media_publish`.
- O `video_url` precisa ser **público e acessível** pela Meta. Subir o arquivo para R2/S3 com URL temporária antes.
- Legenda com crédito no campo `caption`.

## TikTok

- API: Content Posting API.
- Apps não auditados só postam com visibilidade privada. Planejar para isso antes de prometer publicação pública.
- Fluxo: `video/init` → upload em chunks → checar status.

## Agendamento

- Job `publish` no BullMQ com `delay` até `scheduledAt`.
- Antes de publicar, verificar de novo que o render tem origem e licença preenchidas (`smashcut-copyright`).
- Se a publicação falhar, `posts.status = failed` com erro legível e botão de "tentar de novo".

## Testes

- Use contas de teste/privadas para cada rede.
- Para a Meta, use o modo de desenvolvimento do app com contas de teste.
- Não publique conteúdo real em testes automatizados.
