package queue

import (
	"sync"
	"time"

	"github.com/0xrinful/Zenq/internal/models"
)

type JobType int

const (
	JobDownload JobType = iota
	JobOptimize
	JobPack
)

type JobStatus int

const (
	JobPending JobStatus = iota
	JobRunning
	JobDone
	JobFailed
)

type Job struct {
	ID        int
	Type      JobType
	Status    JobStatus
	CreatedAt time.Time
	StartedAt time.Time
	DoneAt    time.Time
	Error     string

	Chapter  models.Chapter
	SrcDir   string
	DestDir  string
	DestFile string
}

type Queue struct {
	downloads chan *Job
	optimizes chan *Job
	packs     chan *Job
	all       []*Job
	nextID    int
	mu        sync.RWMutex
}

func NewQueue() *Queue {
	return &Queue{
		downloads: make(chan *Job, 100),
		optimizes: make(chan *Job, 50),
		packs:     make(chan *Job, 50),
	}
}

func (q *Queue) Enqueue(job *Job) int {
	job.CreatedAt = time.Now().UTC()
	job.Status = JobPending

	q.mu.Lock()
	q.nextID += 1
	job.ID = q.nextID
	q.all = append(q.all, job)
	q.mu.Unlock()

	go func() {
		switch job.Type {
		case JobDownload:
			q.downloads <- job
		case JobOptimize:
			q.optimizes <- job
		case JobPack:
			q.packs <- job
		}
	}()

	return job.ID
}

func (q *Queue) Jobs() []*Job {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.all
}

func (q *Queue) Job(id int) (*Job, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	for _, j := range q.all {
		if j.ID == id {
			return j, true
		}
	}
	return nil, false
}
