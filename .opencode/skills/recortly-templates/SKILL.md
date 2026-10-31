---
name: recortly-templates
description: Use when designing or implementing Recortly templates (profile, zones, phrases, media, style, variations), the template JSON schema, or the engine that turns one clip into many rendered variations.
---

# Recortly: Templates

Um template é um JSON que descreve como montar um post vertical a partir de um clip. O wizard do painel edita esse JSON; o worker o transforma em FFmpeg.

## Etapas do wizard (inspiradas em `img/`)

1. **Perfil**: nome de exibição, @, avatar, selo de verificado.
2. **Zonas**: áreas do vídeo (cabeçalho com perfil, área do vídeo, rodapé).
3. **Frases**: lista de headlines. Uma por variação, ou sorteio.
4. **Mídia**: clip de origem ou upload.
5. **Estilo**: fonte, tamanho, cor, contorno, espaçamento lateral, espaço acima do perfil.
6. **Variações**: quais transformações aplicar e quantas.
7. **Gerar**: cria os renders em fila.

## Schema (exemplo)

```json
{
  "id": "tpl_01",
  "name": "Headline topo",
  "canvas": { "width": 1080, "height": 1920, "fps": 30 },
  "profile": { "name": "...", "handle": "@...", "avatar": "...", "verified": true },
  "zones": {
    "header": { "y": 120, "height": 260 },
    "video":  { "y": 400, "height": 1080, "fit": "cover" },
    "footer": { "y": 1520, "height": 300 }
  },
  "text": {
    "font": "Inter-Bold",
    "size": 53,
    "color": "#FFFFFF",
    "stroke": { "enabled": false, "width": 4, "color": "#000000" },
    "sideSpacing": 84,
    "offsetFromProfile": 156
  },
  "phrases": ["frase 1", "frase 2"],
  "variations": {
    "count": 10,
    "zoom": [1.0, 1.06],
    "saturation": [1.0, 1.1],
    "mirror": false,
    "speed": [0.97, 1.03],
    "seedBase": 42
  }
}
```

## Regras

- Coordenadas sempre no canvas 1080x1920, não em pixels da tela. A UI converte.
- Variações são **determinísticas**: mesma seed + mesmo índice = mesmo resultado.
- Se `phrases` tiver menos itens que `count`, repita em ciclo e varie só a combinação visual.
- Validar o JSON no backend antes de salvar (struct tags + validação explícita). Rejeitar template sem `phrases` ou com zonas fora do canvas.
- Versionar o schema com campo `"schemaVersion": 1`.

## Engine

- `api/internal/video/template.go`: lê o JSON e gera a lista de variações (`[]VariationPlan`).
- `api/internal/video/render.go`: converte cada `VariationPlan` em comando FFmpeg (ver skill `ffmpeg`).
- A lógica de posição fica só no Go. O front Nuxt apenas desenha o preview a partir do JSON.

## Preview

- O preview na UI usa o mesmo JSON. Para a primeira versão, um frame do clip + sobreposição em HTML/CSS basta.
- Mostrar o preview sempre em proporção 9:16 dentro do card, como nas imagens.
