package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/0xrinful/Zenq/internal/models"
	"github.com/0xrinful/Zenq/internal/queue"
)

func (s *Service) Library(ctx context.Context, userID int) ([]models.MangaRecord, error) {
	return s.db.UserMangas(userID)
}

type MangaPageResult struct {
	Manga      *models.MangaRecord
	Chapters   []models.ChapterRecord
	ReadMarks  []float64
	IsFavorite bool
}

func (s *Service) MangaPage(
	ctx context.Context,
	userID int,
	slug, sourceID string,
) (*MangaPageResult, error) {
	manga, err := s.db.Manga(slug, sourceID)
	if err != nil {
		return nil, err
	}
	if manga == nil {
		return nil, ErrNotFound
	}

	chapters, err := s.db.Chapters(slug, sourceID)
	if err != nil {
		return nil, err
	}

	readMarks, err := s.db.ReadMarks(userID, slug, sourceID)
	if err != nil {
		return nil, err
	}

	isFav, err := s.db.IsFavorite(userID, slug, sourceID)
	if err != nil {
		return nil, err
	}

	return &MangaPageResult{
		Manga:      manga,
		Chapters:   chapters,
		ReadMarks:  readMarks,
		IsFavorite: isFav,
	}, nil
}

func (s *Service) SourceLatest(
	ctx context.Context,
	sourceID string,
	page int,
) ([]models.Manga, error) {
	src, ok := s.registry.Source(sourceID)
	if !ok {
		return nil, ErrUnknownSource
	}
	return src.Latest(ctx, page)
}

func (s *Service) SourceSearch(
	ctx context.Context,
	sourceID, query string,
) ([]models.Manga, error) {
	src, ok := s.registry.Source(sourceID)
	if !ok {
		return nil, ErrUnknownSource
	}
	return src.Search(ctx, query)
}

// SourceManga scrapes a manga from a source and displays it
// does not save to DB — just returns the scraped data for viewing
func (s *Service) SourceManga(ctx context.Context, sourceID, slug string) (*models.Manga, error) {
	src, ok := s.registry.Source(sourceID)
	if !ok {
		return nil, ErrUnknownSource
	}
	return src.Manga(ctx, slug)
}

func extractExt(rawURL string) string {
	if i := strings.IndexByte(rawURL, '?'); i != -1 {
		rawURL = rawURL[:i]
	}
	ext := filepath.Ext(rawURL)
	if ext == "" {
		return ".jpg"
	}
	return ext
}

func (s *Service) ImportManga(ctx context.Context, sourceID, slug string) error {
	// scrape and save
	src, ok := s.registry.Source(sourceID)
	if !ok {
		return ErrUnknownSource
	}

	manga, err := src.Manga(ctx, slug)
	if err != nil {
		return fmt.Errorf("service: scrape manga: %w", err)
	}

	// download manga cover
	dl, ok := s.registry.Downloader(manga.SourceID)
	if !ok {
		return ErrUnknownSource
	}

	coverPath := s.files.CoverPath(*manga, extractExt(manga.CoverURL))
	if err := s.files.EnsureDir(filepath.Dir(coverPath)); err != nil {
		return fmt.Errorf("service: ensure cover dir: %w", err)
	}

	err = dl.DownloadCover(ctx, manga.CoverURL, coverPath)
	if err != nil {
		return fmt.Errorf("service: scrape cover: %w", err)
	}

	// save manga record
	err = s.db.SaveManga(models.MangaRecord{
		Manga:     *manga,
		AddedAt:   time.Now(),
		UpdatedAt: time.Now(),
		CoverPath: coverPath,
	})
	if err != nil {
		return fmt.Errorf("service: save manga: %w", err)
	}

	if manga.Chapters == nil {
		chapters, err := src.Chapters(ctx, slug)
		if err != nil {
			return fmt.Errorf("service: scrape chapters: %w", err)
		}
		manga.Chapters = chapters
	}

	// save all chapters
	for _, ch := range manga.Chapters {
		err = s.db.SaveChapter(models.ChapterRecord{Chapter: ch})
		if err != nil {
			return fmt.Errorf("service: save chapter: %w", err)
		}
	}

	return nil
}

// Favorite adds a manga to the user's library
// scrapes the full manga and saves it to DB if not already there
func (s *Service) Favorite(ctx context.Context, userID int, sourceID, slug string) error {
	// check if already in DB
	existing, err := s.db.Manga(slug, sourceID)
	if err != nil {
		return err
	}

	if existing == nil {
		err = s.ImportManga(ctx, sourceID, slug)
		if err != nil {
			return err
		}
	}

	return s.db.AddFavorite(userID, slug, sourceID)
}

func (s *Service) Unfavorite(ctx context.Context, userID int, sourceID, slug string) error {
	return s.db.RemoveFavorite(userID, slug, sourceID)
}

func (s *Service) DownloadChapter(
	ctx context.Context,
	sourceID string,
	chapter models.Chapter,
) (int, error) {
	destDir := s.files.ChapterDir(chapter)
	if err := s.files.EnsureDir(destDir); err != nil {
		return 0, fmt.Errorf("service: ensure dir: %w", err)
	}

	jobID := s.queue.Enqueue(&queue.Job{
		Type:    queue.JobDownload,
		Chapter: chapter,
		DestDir: destDir,
	})

	return jobID, nil
}

func (s *Service) DownloadRange(
	ctx context.Context,
	sourceID, mangaSlug string,
	r models.ChapterRange,
) ([]int, error) {
	chapters, err := s.db.Chapters(mangaSlug, sourceID)
	if err != nil {
		return nil, err
	}

	var jobIDs []int
	for _, ch := range chapters {
		if !r.Contains(ch.Number) {
			continue
		}
		if ch.Downloaded && !r.Force {
			continue
		}

		id, err := s.DownloadChapter(ctx, sourceID, ch.Chapter)
		if err != nil {
			return nil, err
		}
		jobIDs = append(jobIDs, id)
	}

	return jobIDs, nil
}

func (s *Service) OptimizeChapter(ctx context.Context, chapter models.Chapter) (int, error) {
	ch, err := s.db.Chapter(chapter.MangaSlug, chapter.SourceID, chapter.Number)
	if err != nil {
		return 0, err
	}
	if ch == nil || !ch.Downloaded {
		return 0, ErrNotDownloaded
	}

	destDir := s.files.OptimizedDir(chapter)
	if err := s.files.EnsureDir(destDir); err != nil {
		return 0, err
	}

	jobID := s.queue.Enqueue(&queue.Job{
		Type:    queue.JobOptimize,
		Chapter: chapter,
		SrcDir:  ch.RawPath,
		DestDir: destDir,
	})

	return jobID, nil
}

func (s *Service) PackChapter(ctx context.Context, chapter models.Chapter) (int, error) {
	ch, err := s.db.Chapter(chapter.MangaSlug, chapter.SourceID, chapter.Number)
	if err != nil {
		return 0, err
	}
	if ch == nil || !ch.Downloaded {
		return 0, ErrNotDownloaded
	}

	// prefer optimized if available
	srcDir := ch.RawPath
	if ch.Optimized && ch.OptimizedPath != "" {
		srcDir = ch.OptimizedPath
	}

	jobID := s.queue.Enqueue(&queue.Job{
		Type:     queue.JobPack,
		Chapter:  chapter,
		SrcDir:   srcDir,
		DestFile: s.files.CBZPath(chapter),
	})

	return jobID, nil
}

func (s *Service) Jobs() []*queue.Job {
	return s.queue.Jobs()
}

func (s *Service) Job(id int) (*queue.Job, bool) {
	return s.queue.Job(id)
}

func (s *Service) MarkRead(
	ctx context.Context,
	userID int,
	mangaSlug, sourceID string,
	number float64,
) error {
	return s.db.MarkRead(userID, mangaSlug, sourceID, number)
}

func (s *Service) MarkUnread(
	ctx context.Context,
	userID int,
	mangaSlug, sourceID string,
	number float64,
) error {
	return s.db.MarkUnread(userID, mangaSlug, sourceID, number)
}
