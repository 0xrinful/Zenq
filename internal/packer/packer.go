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
	Number  string   `xml:"Number"`
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
