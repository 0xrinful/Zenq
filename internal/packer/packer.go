package packer

import (
	"archive/zip"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/0xrinful/Zenq/internal/models"
)

type Packer struct{}

func New() *Packer {
	return &Packer{}
}

type ComicInfo struct {
	XMLName xml.Name `xml:"ComicInfo"`
	Title   string   `xml:"Title"`
	Series  string   `xml:"Series"`
	Number  string   `xml:"Number"`
	Summary string   `xml:"Summary"`
	Genre   string   `xml:"Genre"`
}

func (p *Packer) Pack(
	ctx context.Context,
	chapter models.Chapter,
	chapterDir, outputPath string,
) error {
	pages, err := os.ReadDir(chapterDir)
	if err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	zw := zip.NewWriter(file)
	defer zw.Close()

	num := strconv.FormatFloat(chapter.Number, 'f', 1, 64)
	num = strings.TrimRight(strings.TrimRight(num, "0"), ".")

	info := ComicInfo{
		Title:  chapter.Title,
		Number: num,
	}

	// fullback
	if strings.TrimSpace(info.Title) == "" {
		info.Title = info.Number
	}

	xmlData, err := xml.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("packer: marshal comic info: %w", err)
	}

	w, err := zw.Create("ComicInfo.xml")
	if err != nil {
		return fmt.Errorf("packer: create ComicInfo.xml: %w", err)
	}

	if _, err := w.Write([]byte(xml.Header)); err != nil {
		return fmt.Errorf("packer: write xml header: %w", err)
	}

	if _, err := w.Write(xmlData); err != nil {
		return fmt.Errorf("packer: write ComicInfo.xml: %w", err)
	}

	for _, entry := range pages {
		if entry.IsDir() {
			continue
		}

		inputPath := filepath.Join(chapterDir, entry.Name())

		src, err := os.Open(inputPath)
		if err != nil {
			return fmt.Errorf("packer: open %s: %w", inputPath, err)
		}

		w, err := zw.Create(entry.Name())
		if err != nil {
			src.Close()
			return fmt.Errorf("packer: create zip entry %s: %w", entry.Name(), err)
		}

		if _, err := io.Copy(w, src); err != nil {
			src.Close()
			return fmt.Errorf("packer: copy %s: %w", entry.Name(), err)
		}

		src.Close()
	}

	return nil
}

func (p *Packer) PackManga(
	ctx context.Context,
	manga models.Manga,
	cbzPaths []string, coverPath, outputPath string,
) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("packer: create output zip: %w", err)
	}
	defer file.Close()

	zw := zip.NewWriter(file)
	defer zw.Close()

	info := ComicInfo{
		Series:  manga.Title,
		Summary: manga.Description,
		Genre:   strings.Join(manga.Genres, ", "),
	}
	xmlData, err := xml.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("packer: marshal manga comic info: %w", err)
	}
	w, err := zw.Create("ComicInfo.xml")
	if err != nil {
		return fmt.Errorf("packer: create ComicInfo.xml: %w", err)
	}
	if _, err := w.Write([]byte(xml.Header)); err != nil {
		return fmt.Errorf("packer: write xml header: %w", err)
	}
	if _, err := w.Write(xmlData); err != nil {
		return fmt.Errorf("packer: write ComicInfo.xml: %w", err)
	}

	if coverPath != "" {
		cover, err := os.Open(coverPath)
		if err != nil {
			return fmt.Errorf("packer: open cover: %w", err)
		}
		coverName := "cover" + filepath.Ext(coverPath)
		w, err := zw.Create(coverName)
		if err != nil {
			cover.Close()
			return fmt.Errorf("packer: create cover entry: %w", err)
		}
		if _, err := io.Copy(w, cover); err != nil {
			cover.Close()
			return fmt.Errorf("packer: copy cover: %w", err)
		}
		cover.Close()
	}

	for _, inputPath := range cbzPaths {
		name := filepath.Base(inputPath)
		src, err := os.Open(inputPath)
		if err != nil {
			return fmt.Errorf("packer: open cbz %s: %w", inputPath, err)
		}
		w, err := zw.Create(name)
		if err != nil {
			src.Close()
			return fmt.Errorf("packer: create zip entry %s: %w", name, err)
		}
		if _, err := io.Copy(w, src); err != nil {
			src.Close()
			return fmt.Errorf("packer: copy cbz %s: %w", name, err)
		}
		src.Close()
	}

	return nil
}
