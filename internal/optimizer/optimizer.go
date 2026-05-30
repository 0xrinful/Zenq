package optimizer

import (
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

const (
	MaxCwebpDim = 16380
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

		if filepath.Ext(inputPath) == ".webp" {
			outputPath := filepath.Join(optimizedDir, fmt.Sprintf("%03d.webp", nextIndex))
			if err := copyFile(inputPath, outputPath); err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", p.Name(), err))
				slog.Error(
					"optimizer failed to copy webp",
					"name", p.Name(),
					"err", err,
				)
				continue
			}

			nextIndex++
			continue
		}

		file, err := os.Open(inputPath)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s (failed to open): %w", p.Name(), err))
			slog.Error("optimizer failed to open", "name", p.Name(), "err", err)
			continue
		}

		imgConfig, _, err := image.DecodeConfig(file)
		file.Close()

		if err != nil {
			errs = append(errs, fmt.Errorf("%s (failed to read dimensions): %w", p.Name(), err))
			slog.Error("optimizer (failed to read dimensions)", "name", p.Name(), "err", err)
			continue
		}

		if imgConfig.Height > MaxCwebpDim || imgConfig.Width > MaxCwebpDim {
			tmpDir, err := os.MkdirTemp("", "zenq_crop_*")
			if err != nil {
				errs = append(
					errs,
					fmt.Errorf("failed to create temp dir for %s: %w", p.Name(), err),
				)
				continue
			}
			defer os.RemoveAll(tmpDir)

			tmpPattern := filepath.Join(tmpDir, "chunk-%03d.jpg")
			cropCmd := exec.CommandContext(ctx,
				"magick", inputPath,
				"-crop", fmt.Sprintf("%dx%d", imgConfig.Width, MaxCwebpDim),
				"+repage",
				tmpPattern,
			)
			slog.Info("crop command", "args", cropCmd.Args)

			out, err := cropCmd.CombinedOutput()
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to crop %s: %w\n%s", p.Name(), err, out))
				slog.Error(
					"optimizer failed to crop",
					"name", p.Name(),
					"err", err,
					"output", string(out),
				)
				continue
			}

			chunks, err := filepath.Glob(filepath.Join(tmpDir, "chunk-*.jpg"))
			if err != nil || len(chunks) == 0 {
				errs = append(errs, fmt.Errorf("no chunks produced for %s", p.Name()))
				slog.Error("optimizer no chunks produced", "name", p.Name())
				continue
			}
			slog.Info("chunks found", "name", p.Name(), "chunks", chunks)

			for _, chunkPath := range chunks {
				outputPath := filepath.Join(optimizedDir, fmt.Sprintf("%03d.webp", nextIndex))
				if err := runCwebp(ctx, chunkPath, outputPath); err != nil {
					errs = append(
						errs,
						fmt.Errorf("%s chunk %s failed: %w", p.Name(), chunkPath, err),
					)
					slog.Error(
						"optimizer failed to run cwebp on chunk",
						"name", p.Name(),
						"chunk", chunkPath,
						"err", err,
					)
				} else {
					nextIndex++
				}
			}

		} else {
			outputPath := filepath.Join(optimizedDir, fmt.Sprintf("%03d.webp", nextIndex))
			if err := runCwebp(ctx, inputPath, outputPath); err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", p.Name(), err))
				slog.Error(
					"optimizer failed to run cwebp",
					"name", p.Name(),
					"err", err,
				)
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

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cwebp failed: %w\n%s", err, out)
	}

	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
