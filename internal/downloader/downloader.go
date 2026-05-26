package downloader

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/0xrinful/Zenq/internal/models"
	"github.com/0xrinful/Zenq/internal/requester"
)

const pageWorkers = 8

type Downloader struct {
	req *requester.Requester
}

func New(req *requester.Requester) *Downloader {
	return &Downloader{req: req}
}

func (d *Downloader) DownloadChapter(
	ctx context.Context,
	chapter models.Chapter,
	destDir string,
) error {
	var (
		mu   sync.Mutex
		wg   sync.WaitGroup
		errs []error
		sem  = make(chan struct{}, pageWorkers)
	)

	for _, page := range chapter.Pages {
		wg.Go(func() {
			sem <- struct{}{}
			defer func() { <-sem }()

			if err := d.downloadPage(ctx, page, destDir); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		})
	}
	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("chapter download had %d page failures", len(errs))
	}
	return nil
}

func (d *Downloader) downloadPage(ctx context.Context, page models.Page, destDir string) error {
	ext := extractExt(page.URL)

	name := fmt.Sprintf("%03d%s", page.Number, ext)
	fullpath := filepath.Join(destDir, name)

	resp, err := d.req.GetImage(ctx, page.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return saveToDisk(resp.Body, fullpath)
}

func (d *Downloader) DownloadCover(ctx context.Context, url string, destPath string) error {
	resp, err := d.req.GetImage(ctx, url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return saveToDisk(resp.Body, destPath)
}

func saveToDisk(body io.Reader, destPath string) error {
	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, body)
	return err
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
