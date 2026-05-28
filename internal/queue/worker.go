package queue

import (
	"context"
	"time"

	"github.com/0xrinful/Zenq/internal/registry"
	"github.com/0xrinful/Zenq/internal/storage/db"
	"github.com/0xrinful/Zenq/internal/storage/files"
)

type Config struct {
	AutoOptimize bool
	AutoPack     bool
}

type Worker struct {
	queue    *Queue
	registry *registry.Registry
	config   Config
	db       *db.DB
	files    *files.Store
}

func NewWorker(
	q *Queue,
	reg *registry.Registry,
	cfg Config,
	db *db.DB,
	files *files.Store,
) *Worker {
	return &Worker{
		queue:    q,
		registry: reg,
		config:   cfg,
		db:       db,
		files:    files,
	}
}

func (w *Worker) Start(ctx context.Context) {
	go w.runPool(ctx, w.queue.downloads, 10)
	go w.runPool(ctx, w.queue.optimizes, 2)
	go w.runPool(ctx, w.queue.packs, 3)
}

func (w *Worker) runPool(ctx context.Context, ch chan *Job, concurrency int) {
	sem := make(chan struct{}, concurrency)
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-ch:
			sem <- struct{}{}
			go func(j *Job) {
				defer func() { <-sem }()
				w.process(ctx, j)
			}(job)
		}
	}
}

func (w *Worker) process(ctx context.Context, job *Job) {
	job.Status = JobRunning
	job.StartedAt = time.Now().UTC()

	var err error
	switch job.Type {
	case JobDownload:
		src, _ := w.registry.Source(job.Chapter.SourceID)

		pages, e := src.Pages(ctx, job.Chapter.URL)
		if e != nil {
			err = e
			break
		}
		job.Chapter.Pages = pages

		dl, _ := w.registry.Downloader(job.Chapter.SourceID)
		err = dl.DownloadChapter(ctx, job.Chapter, job.DestDir)
		if err == nil {
			err = w.db.MarkDownloaded(job.Chapter, job.DestDir)
		}
		if err == nil && w.config.AutoOptimize {
			w.queue.Enqueue(&Job{
				Type:    JobOptimize,
				Chapter: job.Chapter,
				SrcDir:  job.DestDir,
				DestDir: w.files.OptimizedDir(job.Chapter),
			})
		}

	case JobOptimize:
	case JobPack:
	}

	if err != nil {
		job.Status = JobFailed
		job.Error = err.Error()
	} else {
		job.Status = JobDone
	}
	job.DoneAt = time.Now().UTC()
}
