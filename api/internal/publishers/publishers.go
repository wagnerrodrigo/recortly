package publishers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type PublishInput struct {
	VideoPath   string
	Title       string
	Description string
	Visibility  string
}

type PublishResult struct {
	ExternalID  string
	ExternalURL string
}

type Publisher interface {
	Platform() string
	Publish(ctx context.Context, in PublishInput) (PublishResult, error)
}

// YouTube usa a Data API v3 (upload resumable simples de um passo).
type YouTube struct {
	AccessToken string
	HTTP        *http.Client
}

func (y YouTube) Platform() string { return "youtube" }

func (y YouTube) Publish(ctx context.Context, in PublishInput) (PublishResult, error) {
	if strings.TrimSpace(y.AccessToken) == "" {
		return PublishResult{}, fmt.Errorf("conta do YouTube não conectada")
	}
	visibility := in.Visibility
	if visibility == "" {
		visibility = "public"
	}
	file, err := os.Open(in.VideoPath)
	if err != nil {
		return PublishResult{}, err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return PublishResult{}, err
	}

	meta := map[string]any{
		"snippet": map[string]any{
			"title":       truncate(in.Title, 100),
			"description": in.Description,
			"categoryId":  "22",
		},
		"status": map[string]any{
			"privacyStatus":                visibility,
			"selfDeclaredMadeForKids":      false,
		},
	}
	metaRaw, _ := json.Marshal(meta)

	client := y.HTTP
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Minute}
	}

	initReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://www.googleapis.com/upload/youtube/v3/videos?uploadType=resumable&part=snippet,status",
		bytes.NewReader(metaRaw))
	if err != nil {
		return PublishResult{}, err
	}
	initReq.Header.Set("Authorization", "Bearer "+y.AccessToken)
	initReq.Header.Set("Content-Type", "application/json; charset=UTF-8")
	initReq.Header.Set("X-Upload-Content-Type", "video/mp4")
	initReq.Header.Set("X-Upload-Content-Length", fmt.Sprintf("%d", stat.Size()))

	resp, err := client.Do(initReq)
	if err != nil {
		return PublishResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return PublishResult{}, fmt.Errorf("youtube init %d: %s", resp.StatusCode, string(body))
	}
	uploadURL := resp.Header.Get("Location")
	if uploadURL == "" {
		return PublishResult{}, fmt.Errorf("youtube não retornou URL de upload")
	}

	upReq, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, file)
	if err != nil {
		return PublishResult{}, err
	}
	upReq.ContentLength = stat.Size()
	upReq.Header.Set("Content-Type", "video/mp4")

	upResp, err := client.Do(upReq)
	if err != nil {
		return PublishResult{}, err
	}
	defer upResp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(upResp.Body, 4096))
	if upResp.StatusCode >= 300 {
		return PublishResult{}, fmt.Errorf("youtube upload %d: %s", upResp.StatusCode, string(body))
	}
	var out struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(body, &out)
	return PublishResult{
		ExternalID:  out.ID,
		ExternalURL: "https://youtube.com/shorts/" + out.ID,
	}, nil
}

type Unsupported struct {
	Name   string
	Reason string
}

func (u Unsupported) Platform() string { return u.Name }

func (u Unsupported) Publish(context.Context, PublishInput) (PublishResult, error) {
	return PublishResult{}, fmt.Errorf("%s ainda não configurado: %s", u.Name, u.Reason)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
