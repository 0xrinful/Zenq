package optimizer

import (
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

const (
	MaxCwebpDim = 16383
	Quality     = 78
)

type Optimizer struct{}

func (o *Optimizer) OptimizeChapter(ctx context.Context, chapterDir, optimizedDir string) error {
	pages, err := os.ReadDir(chapterDir)
	if err != nil {
		return err
	}

	var errs []error
	nextIndex := 1

	for _, p := range pages {
		inputPath := filepath.Join(chapterDir, p.Name())

		file, err := os.Open(inputPath)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s (failed to open): %w", p.Name(), err))
			continue
		}
		imgConfig, _, err := image.DecodeConfig(file)
		file.Close()

		if err != nil {
			errs = append(errs, fmt.Errorf("%s (failed to read dimensions): %w", p.Name(), err))
			continue
		}

		if imgConfig.Height > MaxCwebpDim || imgConfig.Width > MaxCwebpDim {
			tmpPattern := filepath.Join(os.TempDir(), "zenq_chunk_%03d.jpg")
			cropCmd := exec.CommandContext(ctx,
				"magick", inputPath,
				"-crop",
				fmt.Sprintf("100%%x%d", MaxCwebpDim),
				"+repage",
				tmpPattern,
			)

			if err := cropCmd.Run(); err != nil {
				errs = append(errs, fmt.Errorf("failed to crop large image %s: %w", p.Name(), err))
				continue
			}

			for i := 0; ; i++ {
				chunkPath := fmt.Sprintf(tmpPattern, i)
				if _, err := os.Stat(chunkPath); err != nil {
					break
				}

				outputPath := filepath.Join(optimizedDir, fmt.Sprintf("%03d.webp", nextIndex))
				if err := runCwebp(ctx, chunkPath, outputPath); err != nil {
					errs = append(errs, fmt.Errorf("%s (chunk %d) failed: %w", p.Name(), i, err))
				} else {
					nextIndex++
				}
				os.Remove(chunkPath)
			}

		} else {
			outputPath := filepath.Join(optimizedDir, fmt.Sprintf("%03d.webp", nextIndex))
			if err := runCwebp(ctx, inputPath, outputPath); err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", p.Name(), err))
			} else {
				nextIndex++
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("chapter optimization finished with %d errors", len(errs))
	}
	return nil
}

func runCwebp(ctx context.Context, input, output string) error {
	cmd := exec.CommandContext(ctx,
		"cwebp",
		"-q", strconv.Itoa(Quality),
		"-sharp_yuv",
		"-metadata", "none",
		input,
		"-o", output,
	)
	return cmd.Run()
}
