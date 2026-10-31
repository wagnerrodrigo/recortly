---
name: recortly-copyright
description: Use when handling source videos, licensing or permission fields, credits in captions, or the rule that blocks publishing content without origin data in Recortly.
---

# Recortly: Direitos autorais e créditos

Objetivo: todo vídeo publicado tem origem registrada, crédito visível e possibilidade de remoção rápida.

## Campos obrigatórios em `sources`

- `authorName`: nome do criador original.
- `authorHandle`: @ ou canal.
- `originalUrl`: link da publicação original.
- `licenseType`: um de
  - `own` (é seu)
  - `permission` (permissão explícita do criador)
  - `credit` (uso com crédito, sem permissão formal)
- `permissionProof`: caminho de arquivo ou texto (print, e-mail, data).
- `notes`: observações livres.

## Regras de bloqueio

- Render **não** pode ser publicado se `licenseType` estiver vazio.
- Se `licenseType = credit`, mostrar aviso antes de publicar: "uso com crédito, sem permissão formal".
- Se `licenseType = permission` e `permissionProof` estiver vazio, mostrar aviso.

## Crédito automático

- Legenda/descrição recebe linha no final: `Crédito: {authorName} ({authorHandle}) · {originalUrl}`.
- O crédito é adicionado pelo publisher, não pelo usuário, para não ser esquecido.
- Não remova o crédito ao editar a legenda manualmente.

## Remoção

- Botão "Remover fonte" em cada `source`.
- Ao remover: apagar o arquivo original e todos os renders e arquivos de saída ligados.
- Tentar remover publicações já feitas via API (quando a plataforma permitir) e listar as que precisam de remoção manual.
- Registrar a remoção em `audit_log` (quem, quando, por quê).

## Lembretes para o produto

- Transformar não garante uso legal. Priorize conteúdo próprio ou com permissão.
- Não criar função que baixa conteúdo de perfis alheios sem permissão.
- Se o usuário pedir algo que viole termos das plataformas (scraping, contornar limites, burlar revisão), recusar e sugerir alternativa oficial.
- Não usar marca, logo ou nome de terceiros (incluindo o nome do concorrente) no produto.
