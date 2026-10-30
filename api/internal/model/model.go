package model

import (
	"time"

	"github.com/smashcut/smashcut/api/internal/video/template"
)

type User struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Source struct {
	ID              int64      `json:"id"`
	UserID          int64      `json:"userId"`
	Title           string     `json:"title"`
	FilePath        string     `json:"-"`
	OriginalName    string     `json:"originalName"`
	SizeBytes       int64      `json:"sizeBytes"`
	DurationMS      int64      `json:"durationMs"`
	Width           int        `json:"width"`
	Height          int        `json:"height"`
	HasAudio        bool       `json:"hasAudio"`
	AuthorName      string     `json:"authorName"`
	AuthorHandle    string     `json:"authorHandle"`
	OriginalURL     string     `json:"originalUrl"`
	LicenseType     *string    `json:"licenseType"`
	PermissionProof string     `json:"permissionProof"`
	Notes           string     `json:"notes"`
	CreatedAt       time.Time  `json:"createdAt"`
	MediaURL        string     `json:"mediaUrl,omitempty"`
	Clips           []Clip     `json:"clips,omitempty"`
}

func (s Source) CanPublish() bool {
	if s.LicenseType == nil {
		return false
	}
	switch *s.LicenseType {
	case "own", "permission", "credit":
		return true
	}
	return false
}

type Clip struct {
	ID        int64   `json:"id"`
	SourceID  int64   `json:"sourceId"`
	Name      string  `json:"name"`
	StartMS   int64   `json:"startMs"`
	EndMS     int64   `json:"endMs"`
	Layout    string  `json:"layout"`
	Zoom      float64 `json:"zoom"`
	OffsetX   float64 `json:"offsetX"`
	OffsetY   float64 `json:"offsetY"`
	CreatedAt time.Time `json:"createdAt"`
}

func (c Clip) DurationMS() int64 { return c.EndMS - c.StartMS }

type Template struct {
	ID        int64        `json:"id"`
	UserID    int64        `json:"userId"`
	Name      string       `json:"name"`
	Spec      template.Spec `json:"spec"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
}

type Render struct {
	ID           int64          `json:"id"`
	ClipID       int64          `json:"clipId"`
	TemplateID   *int64         `json:"templateId"`
	VariantCount int            `json:"variantCount"`
	Status       string         `json:"status"`
	Progress     int            `json:"progress"`
	Error        string         `json:"error"`
	CreatedAt    time.Time      `json:"createdAt"`
	FinishedAt   *time.Time     `json:"finishedAt"`
	Outputs      []RenderOutput `json:"outputs,omitempty"`
	Clip         *Clip          `json:"clip,omitempty"`
	SourceID     *int64         `json:"sourceId,omitempty"`
}

type RenderOutput struct {
	ID         int64  `json:"id"`
	RenderID   int64  `json:"renderId"`
	VariantIdx int    `json:"variantIdx"`
	Headline   string `json:"headline"`
	FilePath   string `json:"-"`
	DurationMS int64  `json:"durationMs"`
	MediaURL   string `json:"mediaUrl,omitempty"`
}

type PlatformAccount struct {
	ID             int64      `json:"id"`
	UserID         int64      `json:"userId"`
	Platform       string     `json:"platform"`
	AccountLabel   string     `json:"accountLabel"`
	TokenExpiresAt *time.Time `json:"tokenExpiresAt"`
	Connected      bool       `json:"connected"`
}

type Post struct {
	ID          int64      `json:"id"`
	OutputID    int64      `json:"outputId"`
	Platform    string     `json:"platform"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	ScheduledAt *time.Time `json:"scheduledAt"`
	Status      string     `json:"status"`
	ExternalID  string     `json:"externalId"`
	ExternalURL string     `json:"externalUrl"`
	Error       string     `json:"error"`
	CreatedAt   time.Time  `json:"createdAt"`
	PublishedAt *time.Time `json:"publishedAt"`
	MediaURL    string     `json:"mediaUrl,omitempty"`
}
