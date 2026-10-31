package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client fala com qualquer API compatível com OpenAI
// (LM Studio, Ollama, OpenAI, etc).
type Client struct {
	BaseURL   string
	Model     string
	MaxTokens int
	Reasoning string
	HTTP      *http.Client
}

func New(baseURL, model string) Client {
	return Client{
		BaseURL:   strings.TrimRight(baseURL, "/"),
		Model:     model,
		MaxTokens: 3000,
		HTTP:      &http.Client{Timeout: 180 * time.Second},
	}
}

func (c Client) Configured() bool {
	return c.BaseURL != "" && c.Model != ""
}

type PhrasesRequest struct {
	Idea     string `json:"idea"`
	Count    int    `json:"count"`
	Platform string `json:"platform"`
}

// Phrases gera headlines curtas para o template.
func (c Client) Phrases(ctx context.Context, req PhrasesRequest) ([]string, error) {
	if !c.Configured() {
		return nil, errors.New("IA não configurada: defina RECORTLY_AI_BASE_URL e RECORTLY_AI_MODEL")
	}
	count := req.Count
	if count <= 0 {
		count = 5
	}
	if count > 10 {
		count = 10
	}
	idea := strings.TrimSpace(req.Idea)
	if idea == "" {
		return nil, errors.New("descreva a ideia do vídeo para gerar as frases")
	}

	system := "Você escreve headlines curtas em português brasileiro para vídeos verticais de redes sociais. " +
		"Regras: no máximo 70 caracteres, sem emoji, sem hashtag, tom direto e curioso, sem inventar fatos. " +
		`Responda APENAS um JSON no formato {"phrases":["frase 1","frase 2"]}.`

	user := fmt.Sprintf("Ideia do vídeo: %s\nQuantidade: %d\nPlataforma: %s", idea, count, req.Platform)

	maxTokens := c.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 3000
	}
	body := map[string]any{
		"model": c.Model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
		"temperature": 0.9,
		"max_tokens":  maxTokens,
	}
	// Modelos de raciocínio (gpt-oss) podem gastar todos os tokens pensando.
	if strings.TrimSpace(c.Reasoning) != "" {
		body["reasoning_effort"] = strings.TrimSpace(c.Reasoning)
	}
	raw, _ := json.Marshal(body)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("não foi possível falar com a IA em %s: %w", c.BaseURL, err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("IA respondeu %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("resposta da IA inválida: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return nil, errors.New("a IA não retornou conteúdo")
	}

	var finishReason string
	var parsedFinish struct {
		Choices []struct {
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
	}
	_ = json.Unmarshal(respBody, &parsedFinish)
	if len(parsedFinish.Choices) > 0 {
		finishReason = parsedFinish.Choices[0].FinishReason
	}

	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if content == "" {
		if finishReason == "length" {
			return nil, fmt.Errorf("o modelo gastou todos os tokens em raciocínio antes de escrever as frases. "+
				"Aumente RECORTLY_AI_MAX_TOKENS (atual: %d) ou use um modelo sem raciocínio.", maxTokens)
		}
		return nil, errors.New("a IA respondeu vazio. Verifique o modelo carregado no LM Studio/Ollama")
	}
	phrases := extractPhrases(content)
	if len(phrases) == 0 {
		return nil, errors.New("a IA não retornou frases utilizáveis")
	}
	return phrases, nil
}

// extractPhrases aceita JSON puro ou JSON dentro de bloco markdown.
func extractPhrases(content string) []string {
	content = strings.TrimSpace(content)
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end > start {
		var parsed struct {
			Phrases []string `json:"phrases"`
		}
		if err := json.Unmarshal([]byte(content[start:end+1]), &parsed); err == nil {
			return clean(parsed.Phrases)
		}
	}
	// fallback: uma frase por linha, sem numeração
	lines := strings.Split(content, "\n")
	return clean(lines)
}

func clean(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		s = strings.TrimLeft(s, "-*•0123456789.) ")
		s = strings.Trim(s, `"'`)
		if s == "" || len(s) > 140 {
			continue
		}
		out = append(out, s)
		if len(out) == 10 {
			break
		}
	}
	return out
}
