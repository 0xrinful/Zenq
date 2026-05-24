package sources

import (
	"context"

	"github.com/0xrinful/Zenq/internal/models"
)

type Config struct {
	BaseURL    string
	NeedsFlare bool
	Headers    map[string]string
}

type SourceInfo struct {
	ID   string
	Name string
}

type Source interface {
	Info() SourceInfo
	Latest(ctx context.Context, page int) ([]models.Manga, error)
	Search(ctx context.Context, query string) ([]models.Manga, error)
	Manga(ctx context.Context, slug string) (*models.Manga, error)
	Chapters(ctx context.Context, slug string) ([]models.Chapter, error)
	Pages(ctx context.Context, chapterURL string) ([]models.Page, error)
}
