package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/0xrinful/Zenq/internal/models"
)

type Store struct {
	root string
}

func New(root string) *Store {
	return &Store{root: root}
}

func (s *Store) ChapterDir(chapter models.Chapter) string {
	name := chapterDirName(chapter.Number)
	dir := filepath.Join(s.root, chapter.SourceID, chapter.MangaSlug, name, "raw")
	return dir
}

func (s *Store) OptimizedDir(chapter models.Chapter) string {
	name := chapterDirName(chapter.Number)
	dir := filepath.Join(s.root, chapter.SourceID, chapter.MangaSlug, name, "optimized")
	return dir
}

func (s *Store) CBZPath(chapter models.Chapter) string {
	name := fmt.Sprintf("%s.cbz", chapterDirName(chapter.Number))
	path := filepath.Join(s.root, chapter.SourceID, chapter.MangaSlug, name)
	return path
}

func (s *Store) CBZRange(manga models.Manga, r models.ChapterRange) ([]string, error) {
	mangaDir := filepath.Join(s.root, manga.SourceID, manga.Slug)

	entries, err := os.ReadDir(mangaDir)
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".cbz") {
			continue
		}
		num := parseNumberFromFilename(e.Name())
		if r.Contains(num) {
			paths = append(paths, filepath.Join(mangaDir, e.Name()))
		}
	}

	return paths, nil
}

func (s *Store) CoverPath(manga models.Manga, ext string) string {
	path := filepath.Join(s.root, manga.SourceID, manga.Slug, "cover"+ext)
	return path
}

func (s *Store) EnsureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

func chapterDirName(number float64) string {
	return fmt.Sprintf("chapter-%07.3f", number)
}

func parseNumberFromFilename(name string) float64 {
	name = strings.TrimSuffix(name, filepath.Ext(name))
	name = strings.TrimPrefix(name, "chapter-")
	number, _ := strconv.ParseFloat(name, 64)
	return number
}
