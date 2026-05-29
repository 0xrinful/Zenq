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

func (s *Store) Root() string {
	return s.root
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
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".cbz") {
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

func (s *Store) ResolvePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("files: empty path")
	}

	root, err := filepath.Abs(s.root)
	if err != nil {
		return "", fmt.Errorf("files: resolve root: %w", err)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("files: resolve path: %w", err)
	}

	if absPath != root && !strings.HasPrefix(absPath, root+string(os.PathSeparator)) {
		return "", fmt.Errorf("files: path outside root")
	}

	return absPath, nil
}

func (s *Store) ResolveFile(dir, name string) (string, error) {
	if name == "" || name == "." || name == ".." || filepath.Base(name) != name {
		return "", fmt.Errorf("files: invalid filename")
	}

	base, err := s.ResolvePath(dir)
	if err != nil {
		return "", err
	}

	path := filepath.Join(base, name)
	if path != base && !strings.HasPrefix(path, base+string(os.PathSeparator)) {
		return "", fmt.Errorf("files: invalid file path")
	}

	return path, nil
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
