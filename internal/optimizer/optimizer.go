package optimizer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const (
	Quality     = 78
	pageWorkers = 1
)

type Optimizer struct{}

func New() *Optimizer {
	return &Optimizer{}
}

func (o *Optimizer) OptimizeChapter(ctx context.Context, chapterDir, optimizedDir string) error {
	pages, err := os.ReadDir(chapterDir)
	if err != nil {
		return err
	}

	var (
		mu   sync.Mutex
		wg   sync.WaitGroup
		sema = make(chan struct{}, pageWorkers)
		errs []error
	)

	for _, p := range pages {
		sema <- struct{}{}
		wg.Go(func() {
			defer func() { <-sema }()
			inputPath := filepath.Join(chapterDir, p.Name())
			outputName := filepath.Base(p.Name())
			outputName = strings.TrimSuffix(outputName, filepath.Ext(outputName)) + ".webp"
			outputPath := filepath.Join(optimizedDir, outputName)
			cmd := exec.CommandContext(ctx,
				"cwebp",
				"-q", strconv.Itoa(Quality),
				"-sharp_yuv",
				"-metadata", "none",
				inputPath,
				"-o", outputPath,
			)

			if err := cmd.Run(); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("%s: %w", p.Name(), err))
				mu.Unlock()
			}
		})
	}
	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("chapter optimization failed: %d pages failed", len(errs))
	}
	return nil
}
