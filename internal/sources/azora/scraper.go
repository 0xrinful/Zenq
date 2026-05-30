package azora

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/0xrinful/Zenq/internal/models"
)

func (s *Source) Latest(ctx context.Context, page, size int) ([]models.Manga, error) {
	return s.queryMangas(ctx, "", page, size)
}

func (s *Source) Search(ctx context.Context, query string, page, size int) ([]models.Manga, error) {
	return s.queryMangas(ctx, query, page, size)
}

func (s *Source) queryMangas(
	ctx context.Context,
	query string,
	page, size int,
) ([]models.Manga, error) {
	u := fmt.Sprintf(
		"%s/api/posts?page=%d&perPage=%d&searchTerm=%s&tag=hot",
		apiURL,
		page,
		size,
		url.QueryEscape(query),
	)

	resp, err := s.req.Get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("azora: query: %w", err)
	}
	defer resp.Body.Close()

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("azora: parse query: %w", err)
	}

	var mangas []models.Manga
	for _, p := range result.Posts {
		if p.SeriesType == "NOVEL" {
			continue
		}
		mangas = append(mangas, models.Manga{
			Slug:     p.Slug,
			SourceID: sourceID,
			Title:    strings.TrimSpace(p.Title),
			CoverURL: p.FeaturedImage,
		})
	}

	return mangas, nil
}

func (s *Source) Manga(ctx context.Context, slug string) (*models.Manga, error) {
	u := fmt.Sprintf("%s/api/post?postSlug=%s", apiURL, url.QueryEscape(slug))

	resp, err := s.req.Get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("azora: manga: %w", err)
	}
	defer resp.Body.Close()

	var result postResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("azora: parse manga: %w", err)
	}

	p := result.Post

	manga := &models.Manga{
		Slug:        p.Slug,
		SourceID:    sourceID,
		Title:       strings.TrimSpace(p.Title),
		Description: strings.TrimSpace(stripHTML(p.Description)),
		CoverURL:    p.FeaturedImage,
		Status:      p.SeriesStatus,
	}

	for _, g := range p.Genres {
		manga.Genres = append(manga.Genres, g.Name)
	}

	for _, c := range p.Chapters {
		// Only include free, accessible chapters.
		if c.IsLocked || !c.IsAccessible {
			continue
		}

		releasedAt, _ := time.Parse(time.RFC3339, c.CreatedAt)

		manga.Chapters = append(manga.Chapters, models.Chapter{
			SourceID:  sourceID,
			MangaSlug: slug,
			Title:     strings.TrimSpace(c.Title),
			Number:    c.Number,
			// Store the chapter ID directly so Pages() can call /api/chapter
			// without any extra resolution requests.
			URL:        fmt.Sprintf("%d", c.ID),
			ReleasedAt: releasedAt.UTC(),
		})
	}

	return manga, nil
}

func (s *Source) Chapters(ctx context.Context, slug string) ([]models.Chapter, error) {
	manga, err := s.Manga(ctx, slug)
	if err != nil {
		return nil, err
	}
	return manga.Chapters, nil
}

func (s *Source) Pages(ctx context.Context, chapterURL string) ([]models.Page, error) {
	// chapterURL is the raw chapter ID stored by Manga() / Chapters().
	u := fmt.Sprintf("%s/api/chapter?chapterId=%s", apiURL, chapterURL)

	resp, err := s.req.Get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("azora: pages: %w", err)
	}
	defer resp.Body.Close()

	var result chapterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("azora: parse pages: %w", err)
	}

	ch := result.Chapter

	if ch.IsShortLinkLocked {
		return nil, fmt.Errorf("azora: chapter locked (short link)")
	}
	if ch.IsLockedByCoins {
		return nil, fmt.Errorf("azora: chapter locked (coins required)")
	}
	if ch.IsPermanentlyLocked {
		return nil, fmt.Errorf("azora: chapter permanently locked")
	}

	images := ch.Images
	sort.Slice(images, func(i, j int) bool {
		return images[i].Order < images[j].Order
	})

	pages := make([]models.Page, 0, len(images))
	for i, img := range images {
		pages = append(pages, models.Page{
			Number: i + 1,
			URL:    strings.ReplaceAll(img.URL, " ", "%20"),
		})
	}

	return pages, nil
}

func stripHTML(raw string) string {
	var result strings.Builder
	result.Grow(len(raw))

	inTag := false
	for i, r := range raw {
		switch {
		case inTag && r == 'p' && raw[i-1] == '/':
			result.WriteByte('\n')
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag && r != '\\':
			result.WriteRune(r)
		}
	}

	return result.String()
}
