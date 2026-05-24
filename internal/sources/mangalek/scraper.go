package mangalek

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/0xrinful/Zenq/internal/helpers/parse"
	"github.com/0xrinful/Zenq/internal/models"
)

func (s *Source) Latest(ctx context.Context, page int) ([]models.Manga, error) {
	payload := fmt.Sprintf(
		"action=madara_load_more&page=%d&template=madara-core%%2Fcontent%%2Fcontent-archive&vars%%5Bpost_type%%5D=wp-manga&vars%%5Bpost_status%%5D=publish&vars%%5Bmeta_key%%5D=_latest_update&vars%%5Borderby%%5D=meta_value_num&vars%%5Border%%5D=desc",
		page-1,
	)

	ajaxURL := fmt.Sprintf("%s/wp-admin/admin-ajax.php", baseURL)

	resp, err := s.req.Post(ctx, ajaxURL, "application/x-www-form-urlencoded", []byte(payload))
	if err != nil {
		return nil, fmt.Errorf("mangalek: latest: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("mangalek: parse latest: %w", err)
	}

	var mangas []models.Manga

	doc.Find("div.page-item-detail.manga").Each(func(i int, s *goquery.Selection) {
		a := s.Find(".post-title h3 a")
		mangaURL := a.AttrOr("href", "")

		slug := strings.TrimSuffix(mangaURL, "/")
		if idx := strings.LastIndex(slug, "/"); idx != -1 {
			slug = slug[idx+1:]
		}

		img := s.Find(".item-thumb img")
		coverURL := img.AttrOr("src", "")
		if srcset := img.AttrOr("srcset", ""); srcset != "" {
			coverURL = strings.Fields(srcset[strings.LastIndex(srcset, ",")+1:])[0]
		}

		manga := models.Manga{
			Slug:     slug,
			SourceID: sourceID,
			Title:    strings.TrimSpace(a.Text()),
			CoverURL: coverURL,
		}

		mangas = append(mangas, manga)
	})

	return mangas, nil
}

func (s *Source) Search(ctx context.Context, query string) ([]models.Manga, error) {
	payload := fmt.Sprintf(
		"action=madara_load_more&page=0&template=madara-core%%2Fcontent%%2Fcontent-archive&vars%%5Bpost_type%%5D=wp-manga&vars%%5Bpost_status%%5D=publish&vars%%5Bs%%5D=%s",
		url.QueryEscape(query),
	)

	ajaxURL := fmt.Sprintf("%s/wp-admin/admin-ajax.php", baseURL)

	resp, err := s.req.Post(ctx, ajaxURL, "application/x-www-form-urlencoded", []byte(payload))
	if err != nil {
		return nil, fmt.Errorf("mangalek: search: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("mangalek: parse search: %w", err)
	}

	var mangas []models.Manga

	doc.Find("div.page-item-detail.manga").Each(func(i int, s *goquery.Selection) {
		a := s.Find(".post-title h3 a")
		mangaURL := a.AttrOr("href", "")
		if mangaURL == "" {
			return
		}

		slug := strings.TrimSuffix(mangaURL, "/")
		if idx := strings.LastIndex(slug, "/"); idx != -1 {
			slug = slug[idx+1:]
		}

		img := s.Find(".item-thumb img")
		coverURL := img.AttrOr("src", "")
		if srcset := img.AttrOr("srcset", ""); srcset != "" {
			coverURL = strings.Fields(srcset[strings.LastIndex(srcset, ",")+1:])[0]
		}

		mangas = append(mangas, models.Manga{
			Slug:     slug,
			SourceID: sourceID,
			Title:    strings.TrimSpace(a.Text()),
			CoverURL: coverURL,
		})
	})

	return mangas, nil
}

func (s *Source) Manga(ctx context.Context, slug string) (*models.Manga, error) {
	url := fmt.Sprintf("%s/manga/%s", baseURL, slug)

	resp, err := s.req.Get(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("mangalek: manga: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("mangalek: parse manga: %w", err)
	}

	manga := &models.Manga{
		Slug:     slug,
		SourceID: sourceID,
	}

	manga.Title = strings.TrimSpace(doc.Find(".post-title h1").Text())
	manga.Description = strings.TrimSpace(doc.Find(".summary__content p").Text())
	manga.CoverURL = doc.Find(".summary_image img").AttrOr("src", "")
	manga.Status = strings.TrimSpace(doc.Find(".summary-content").Last().Text())

	doc.Find("div.genres-content a").Each(func(i int, s *goquery.Selection) {
		manga.Genres = append(manga.Genres, strings.TrimSpace(s.Text()))
	})

	doc.Find("li.wp-manga-chapter").Each(func(i int, s *goquery.Selection) {
		chapter := models.Chapter{}
		a := s.Find("a")
		chapter.URL = a.AttrOr("href", "")

		title := strings.TrimSpace(a.Text())
		number, _ := parse.ParseChapterNumber(title)
		chapter.Title = title
		chapter.Number = number

		dateText := strings.TrimSpace(
			s.Find(".chapter-release-date i").Text(),
		)

		if dateText == "" {
			dateText = s.Find(".c-new-tag a").AttrOr("title", "")
		}

		releasedAt, _ := parse.ParseDate(dateText)
		chapter.ReleasedAt = releasedAt.UTC()

		manga.Chapters = append(manga.Chapters, chapter)
	})

	return manga, nil
}

func (s *Source) Chapters(ctx context.Context, slug string) ([]models.Chapter, error) {
	return nil, nil
}

func (s *Source) Pages(ctx context.Context, chapterURL string) ([]models.Page, error) {
	resp, err := s.req.Get(ctx, chapterURL)
	if err != nil {
		return nil, fmt.Errorf("mangalek: manga: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("mangalek: parse manga: %w", err)
	}

	var pages []models.Page

	doc.Find("img.wp-manga-chapter-img").Each(func(i int, s *goquery.Selection) {
		pages = append(pages, models.Page{
			Number: i + 1,
			URL:    s.AttrOr("src", ""),
		})
	})

	return pages, nil
}
