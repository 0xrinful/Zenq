package handlers

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"

	"github.com/0xrinful/Zenq/internal/queue"
	"github.com/0xrinful/Zenq/internal/service"
)

type Dashboard struct {
	svc       *service.Service
	templates map[string]*template.Template
}

type jobDesc struct {
	ID          int
	Description string
	Status      queue.JobStatus
	CreatedAt   time.Time
	StartedAt   *time.Time
	DoneAt      *time.Time
	Error       string
}

type jobCounts struct {
	All     int
	Pending int
	Running int
	Done    int
	Failed  int
	Active  string
}

type dashboardPageData struct {
	CurrentPath  string
	MangaCount   int
	RunningCount int
	Counts       jobCounts
	InitialJobs  []jobDesc
}

type storagePartialData struct {
	UsedBytes  int64
	TotalBytes uint64
	Used       string
	Total      string
	Percent    float64
}

func NewDashboard(svc *service.Service, templates map[string]*template.Template) *Dashboard {
	return &Dashboard{svc: svc, templates: templates}
}

func buildDescription(j *queue.Job) string {
	if j == nil {
		return "Unknown job"
	}

	slug := titleSlug(j.Chapter.MangaSlug)
	num := strconv.FormatFloat(j.Chapter.Number, 'f', -1, 64)

	switch j.Type {
	case queue.JobDownload:
		return fmt.Sprintf("Downloading %s ch.%s", slug, num)
	case queue.JobOptimize:
		return fmt.Sprintf("Optimizing %s ch.%s", slug, num)
	case queue.JobPack:
		return fmt.Sprintf("Packing %s ch.%s", slug, num)
	default:
		return fmt.Sprintf("Processing %s ch.%s", slug, num)
	}
}

func (d *Dashboard) Page(w http.ResponseWriter, r *http.Request) {
	jobs := d.svc.Jobs()
	counts := countJobs(jobs, "all")
	mangas, err := d.svc.Library(r.Context(), getUserID(r.Context()))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderTemplate(w, d.templates, "dashboard.html", dashboardPageData{
		CurrentPath:  "dashboard",
		MangaCount:   len(mangas),
		RunningCount: counts.Running,
		Counts:       counts,
		InitialJobs:  describeJobs(jobs),
	})
}

func (d *Dashboard) Jobs(w http.ResponseWriter, r *http.Request) {
	jobs := d.svc.Jobs()
	status := strings.TrimSpace(r.URL.Query().Get("status"))

	if r.URL.Query().Get("count") != "" {
		counts := countJobs(jobs, status)
		renderTemplateName(w, d.templates, "dashboard.html", "job-count-partial", counts)
		return
	}

	filtered := filterJobsByStatus(jobs, status)
	renderTemplateName(
		w,
		d.templates,
		"dashboard.html",
		"jobs-list-partial",
		describeJobs(filtered),
	)
}

func (d *Dashboard) JobDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid job id", http.StatusBadRequest)
		return
	}

	job, ok := d.svc.Job(id)
	if !ok {
		http.NotFound(w, r)
		return
	}

	renderTemplateName(w, d.templates, "dashboard.html", "job-detail-partial", describeJob(job))
}

func (d *Dashboard) Storage(w http.ResponseWriter, r *http.Request) {
	root := d.svc.Files().Root()
	used, err := usedBytes(root)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var total uint64
	var stat syscall.Statfs_t
	if err := syscall.Statfs(root, &stat); err == nil {
		total = stat.Blocks * uint64(stat.Bsize)
	}

	data := storagePartialData{
		UsedBytes:  used,
		TotalBytes: total,
		Used:       formatBytes(uint64(used)),
		Total:      formatOptionalBytes(total),
	}
	if total > 0 {
		data.Percent = (float64(used) / float64(total)) * 100
	}

	renderTemplateName(w, d.templates, "dashboard.html", "storage-partial", data)
}

func (d *Dashboard) StartFlareSolver(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command(
		"docker", "run", "-d", "--rm", "-p", "8191:8191",
		"--name", "flaresolverr",
		"ghcr.io/flaresolverr/flaresolverr:latest",
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		writeToast(w, message, "error")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	writeToast(w, "FlareSolver started", "success")
	renderTemplateName(w, d.templates, "dashboard.html", "flare-status-partial", nil)
}

func describeJobs(jobs []*queue.Job) []jobDesc {
	desc := make([]jobDesc, 0, len(jobs))
	for i := len(jobs) - 1; i >= 0; i-- {
		desc = append(desc, describeJob(jobs[i]))
	}
	return desc
}

func describeJob(job *queue.Job) jobDesc {
	desc := jobDesc{
		ID:          job.ID,
		Description: buildDescription(job),
		Status:      job.Status,
		CreatedAt:   job.CreatedAt,
		Error:       job.Error,
	}
	if !job.StartedAt.IsZero() {
		desc.StartedAt = &job.StartedAt
	}
	if !job.DoneAt.IsZero() {
		desc.DoneAt = &job.DoneAt
	}
	return desc
}

func filterJobsByStatus(jobs []*queue.Job, rawStatus string) []*queue.Job {
	status, ok := parseJobStatus(rawStatus)
	if !ok {
		return jobs
	}

	filtered := make([]*queue.Job, 0, len(jobs))
	for _, job := range jobs {
		if job.Status == status {
			filtered = append(filtered, job)
		}
	}
	return filtered
}

func countJobs(jobs []*queue.Job, active string) jobCounts {
	counts := jobCounts{All: len(jobs), Active: active}
	if counts.Active == "" {
		counts.Active = "all"
	}
	for _, job := range jobs {
		switch job.Status {
		case queue.JobPending:
			counts.Pending++
		case queue.JobRunning:
			counts.Running++
		case queue.JobDone:
			counts.Done++
		case queue.JobFailed:
			counts.Failed++
		}
	}
	return counts
}

func parseJobStatus(raw string) (queue.JobStatus, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "pending":
		return queue.JobPending, true
	case "running":
		return queue.JobRunning, true
	case "done":
		return queue.JobDone, true
	case "failed":
		return queue.JobFailed, true
	default:
		return queue.JobPending, false
	}
}

func titleSlug(slug string) string {
	words := strings.Fields(strings.ReplaceAll(slug, "-", " "))
	for i, word := range words {
		words[i] = titleWord(word)
	}
	return strings.Join(words, " ")
}

func titleWord(word string) string {
	if word == "" {
		return word
	}
	runes := []rune(strings.ToLower(word))
	runes[0] = unicode.ToTitle(runes[0])
	return string(runes)
}

func usedBytes(root string) (int64, error) {
	var total int64
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return err
		}
		if info == nil || info.IsDir() {
			return nil
		}
		total += info.Size()
		return nil
	})
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	return total, err
}

func formatBytes(bytes uint64) string {
	if bytes == 0 {
		return "0 B"
	}
	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	value := float64(bytes)
	unit := 0
	for value >= 1024 && unit < len(units)-1 {
		value /= 1024
		unit++
	}
	if unit == 0 {
		return fmt.Sprintf("%d %s", bytes, units[unit])
	}
	return fmt.Sprintf("%.1f %s", value, units[unit])
}

func formatOptionalBytes(bytes uint64) string {
	if bytes == 0 {
		return "—"
	}
	return formatBytes(bytes)
}
