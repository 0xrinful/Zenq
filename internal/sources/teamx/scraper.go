package teamx

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/0xrinful/Zenq/internal/models"
)

// ─── Latest ──────────────────────────────────────────────────────────────────
// The home page lists recently updated series inside "div.last-chapter div.box".
// Pages are paginated with <a rel="next">.

func (s *Source) Latest(ctx context.Context, page, size int) ([]models.Manga, error) {
	url := baseURL
	if page > 1 {
		url = fmt.Sprintf("%s?page=%d", baseURL, page)
	}

	resp, err := s.req.Get(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("teamx: latest: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("teamx: parse latest: %w", err)
	}

	seen := make(map[string]struct{})
	var mangas []models.Manga

	doc.Find("div.last-chapter div.box").Each(func(_ int, sel *goquery.Selection) {
		link := sel.Find("div.info a")
		title := strings.TrimSpace(link.Find("h3").Text())
		if title == "" || len(mangas) >= size {
			return
		}
		if _, dup := seen[title]; dup {
			return
		}
		seen[title] = struct{}{}

		href := link.AttrOr("href", "")
		slug := extractSlug(href)

		coverURL := sel.Find("div.imgu img").AttrOr("src", "")
		imgName := strings.TrimPrefix(coverURL, baseURL+"/images/manga/thumbnail_")
		coverURL = baseURL + "/images/manga/" + imgName

		mangas = append(mangas, models.Manga{
			Slug:     slug,
			SourceID: sourceID,
			Title:    title,
			CoverURL: coverURL,
		})
	})

	return mangas, nil
}

// ─── Search ───────────────────────────────────────────────────────────────────
// TeamX exposes an AJAX search endpoint that returns an HTML fragment.
// For browse-without-query we hit /series?page=N (same HTML as popular).

func (s *Source) Search(ctx context.Context, query string, page, size int) ([]models.Manga, error) {
	if strings.TrimSpace(query) != "" {
		return s.ajaxSearch(ctx, query)
	}
	return s.browseSeries(ctx, page, size)
}

func (s *Source) ajaxSearch(ctx context.Context, query string) ([]models.Manga, error) {
	url := fmt.Sprintf("%s/ajax/search?keyword=%s", baseURL, encodeQuery(query))

	resp, err := s.req.Get(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("teamx: search: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("teamx: parse search: %w", err)
	}

	var mangas []models.Manga

	// AJAX results: <a class="items-center" href="..."><img ...><h4>title</h4></a>
	doc.Find("a.items-center").Each(func(_ int, sel *goquery.Selection) {
		title := strings.TrimSpace(sel.Find("h4").Text())
		if title == "" {
			return
		}
		href := sel.AttrOr("href", "")
		slug := extractSlug(href)
		coverURL := sel.Find("img").AttrOr("src", "")
		imgName := strings.TrimPrefix(coverURL, baseURL+"/images/manga/thumbnail_")
		coverURL = baseURL + "/images/manga/" + imgName

		mangas = append(mangas, models.Manga{
			Slug:     slug,
			SourceID: sourceID,
			Title:    title,
			CoverURL: coverURL,
		})
	})

	return mangas, nil
}

func (s *Source) browseSeries(ctx context.Context, page, size int) ([]models.Manga, error) {
	url := fmt.Sprintf("%s/series?page=%d", baseURL, page)

	resp, err := s.req.Get(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("teamx: browse: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("teamx: parse browse: %w", err)
	}

	var mangas []models.Manga

	// /series grid: div.listupd div.bsx
	doc.Find("div.listupd div.bsx").Each(func(_ int, sel *goquery.Selection) {
		if len(mangas) >= size {
			return
		}
		a := sel.Find("a").First()
		title := strings.TrimSpace(a.AttrOr("title", ""))
		href := a.AttrOr("href", "")
		slug := extractSlug(href)

		img := sel.Find("img")
		coverURL := img.AttrOr("data-src", img.AttrOr("src", ""))

		mangas = append(mangas, models.Manga{
			Slug:     slug,
			SourceID: sourceID,
			Title:    title,
			CoverURL: coverURL,
		})
	})

	return mangas, nil
}

// ─── Manga detail ─────────────────────────────────────────────────────────────

func (s *Source) Manga(ctx context.Context, slug string) (*models.Manga, error) {
	url := fmt.Sprintf("%s/series/%s", baseURL, slug)

	resp, err := s.req.Get(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("teamx: manga: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("teamx: parse manga: %w", err)
	}

	manga := &models.Manga{
		Slug:     slug,
		SourceID: sourceID,
	}

	manga.Title = strings.TrimSpace(doc.Find("div.author-info-title h1").Text())

	// Description: prefer the plain text container, fall back to <p> children.
	desc := strings.TrimSpace(doc.Find("div.review-content").Text())
	if desc == "" {
		doc.Find("div.review-content p").Each(func(_ int, sel *goquery.Selection) {
			desc += strings.TrimSpace(sel.Text()) + "\n"
		})
		desc = strings.TrimSpace(desc)
	}
	manga.Description = desc

	manga.CoverURL = doc.Find("div.text-right img").First().AttrOr("src", "")

	// Status: the element after the <small> that contains "الحالة"
	doc.Find(".full-list-info > small").Each(func(_ int, sel *goquery.Selection) {
		if strings.Contains(sel.Text(), "الحالة") {
			manga.Status = strings.TrimSpace(sel.Next().Text())
		}
	})

	// Genres
	doc.Find("div.review-author-info a").Each(func(_ int, sel *goquery.Selection) {
		if g := strings.TrimSpace(sel.Text()); g != "" {
			manga.Genres = append(manga.Genres, g)
		}
	})

	// Chapters — may be paginated; follow <a rel="next"> until exhausted.
	manga.Chapters, err = s.fetchAllChapters(ctx, slug, doc)
	if err != nil {
		return nil, err
	}

	return manga, nil
}

// fetchAllChapters collects chapters from the first page doc and follows
// pagination links until there are no more pages.
func (s *Source) fetchAllChapters(
	ctx context.Context,
	slug string,
	firstDoc *goquery.Document,
) ([]models.Chapter, error) {
	var chapters []models.Chapter
	doc := firstDoc

	for {
		doc.Find("div.chapter-card").Each(func(_ int, sel *goquery.Selection) {
			ch := parseChapterCard(sel, slug)
			if ch.URL != "#" {
				chapters = append(chapters, ch)
			}
		})

		next := doc.Find("a[rel=next]").AttrOr("href", "")
		if next == "" {
			break
		}

		resp, err := s.req.Get(ctx, next)
		if err != nil {
			return chapters, fmt.Errorf("teamx: chapters page: %w", err)
		}

		doc, err = goquery.NewDocumentFromReader(resp.Body)
		resp.Body.Close()
		if err != nil {
			return chapters, fmt.Errorf("teamx: parse chapters page: %w", err)
		}
	}

	return chapters, nil
}

func parseChapterCard(sel *goquery.Selection, slug string) models.Chapter {
	numberStr := sel.AttrOr("data-number", "")
	number, _ := strconv.ParseFloat(numberStr, 64)

	titleText := strings.TrimSpace(
		sel.Find("div.chapter-info div.chapter-title").Text(),
	)
	title := fmt.Sprintf("الفصل %s", numberStr)
	if titleText != "" &&
		titleText != numberStr &&
		titleText != title &&
		titleText != fmt.Sprintf("الفصل رقم %s", numberStr) {
		title = fmt.Sprintf("%s - %s", title, titleText)
	}

	// data-date is a Unix timestamp in seconds.
	var releasedAt time.Time
	if ts, err := strconv.ParseInt(sel.AttrOr("data-date", ""), 10, 64); err == nil && ts > 0 {
		releasedAt = time.Unix(ts, 0).UTC()
	}

	chURL := sel.Find("a").AttrOr("href", "")
	// chURL := stripDomain(href, baseURL)

	return models.Chapter{
		SourceID:   sourceID,
		MangaSlug:  slug,
		Title:      title,
		Number:     number,
		URL:        chURL,
		ReleasedAt: releasedAt,
	}
}

// ─── Chapters ─────────────────────────────────────────────────────────────────

func (s *Source) Chapters(ctx context.Context, slug string) ([]models.Chapter, error) {
	manga, err := s.Manga(ctx, slug)
	if err != nil {
		return nil, err
	}
	return manga.Chapters, nil
}

// ─── Pages ────────────────────────────────────────────────────────────────────

func (s *Source) Pages(ctx context.Context, chapterURL string) ([]models.Page, error) {
	resp, err := s.req.Get(ctx, chapterURL)
	if err != nil {
		return nil, fmt.Errorf("teamx: pages: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("teamx: parse pages: %w", err)
	}

	var pages []models.Page

	// Images are inside div.image_list, either <canvas data-src> or <img src>.
	doc.Find("div.image_list canvas[data-src], div.image_list img[src]").
		Each(func(i int, sel *goquery.Selection) {
			imgURL := sel.AttrOr("src", sel.AttrOr("data-src", ""))
			if imgURL == "" {
				return
			}
			// Make absolute if relative.
			if !strings.HasPrefix(imgURL, "http") {
				imgURL = baseURL + imgURL
			}
			pages = append(pages, models.Page{
				Number: i + 1,
				URL:    imgURL,
			})
		})

	return pages, nil
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// extractSlug returns the last non-empty path segment of a URL string.
func extractSlug(href string) string {
	href = strings.TrimSuffix(href, "/")
	if idx := strings.LastIndex(href, "/"); idx != -1 {
		return href[idx+1:]
	}
	return href
}

// encodeQuery percent-encodes a search query for use in a URL.
func encodeQuery(q string) string {
	var sb strings.Builder
	for _, b := range []byte(q) {
		switch {
		case b >= 'A' && b <= 'Z',
			b >= 'a' && b <= 'z',
			b >= '0' && b <= '9',
			b == '-', b == '_', b == '.', b == '~':
			sb.WriteByte(b)
		default:
			fmt.Fprintf(&sb, "%%%02X", b)
		}
	}
	return sb.String()
}
