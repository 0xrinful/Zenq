package server

import (
	"fmt"
	"html/template"
	"log/slog"
	"math/rand"
	"net/http"
	"path/filepath"
	"reflect"
	"sync"

	"github.com/0xrinful/Zenq/internal/models"
	"github.com/0xrinful/Zenq/internal/queue"
	"github.com/0xrinful/Zenq/internal/service"
)

var (
	muxMu         sync.Mutex
	registeredMux http.Handler
	sessionSecret string
)

type ActionSection struct {
	Label      string
	Endpoint   string
	HasRange   bool
	IsDelete   bool
	IsDownload bool
}

type ChapterRowData struct {
	SourceID  string
	MangaSlug string
	Chapter   models.ChapterRecord
	IsRead    bool
}

type jobFilter struct {
	Status string
	Label  string
	Count  int
	Active bool
}

type Server struct {
	service   *service.Service
	templates map[string]*template.Template
}

func SetSessionSecret(secret string) {
	muxMu.Lock()
	sessionSecret = secret
	registeredMux = nil
	muxMu.Unlock()
}

func actionSections(sourceID, mangaSlug string) []ActionSection {
	base := fmt.Sprintf("/manga/%s/%s", sourceID, mangaSlug)
	return []ActionSection{
		{Label: "Download", Endpoint: base + "/download", HasRange: true},
		{Label: "Optimize", Endpoint: base + "/optimize", HasRange: true},
		{Label: "Pack", Endpoint: base + "/pack", HasRange: true},
		{
			Label:      "Download as ZIP",
			Endpoint:   base + "/download/zip",
			HasRange:   true,
			IsDownload: true,
		},
		{Label: "Chapters", Endpoint: base + "/refresh"},
		{Label: "Delete Files", Endpoint: base + "/files", IsDelete: true},
	}
}

func chapterRowData(
	sourceID string,
	mangaSlug string,
	chapter models.ChapterRecord,
	readMarks map[float64]bool,
) ChapterRowData {
	return ChapterRowData{
		SourceID:  sourceID,
		MangaSlug: mangaSlug,
		Chapter:   chapter,
		IsRead:    readMarks[chapter.Number],
	}
}

func jobStatusClass(status queue.JobStatus) string {
	switch status {
	case queue.JobRunning:
		return "animate-glow-pulse border-l-2 border-neon-blue bg-neon-blue/5"
	case queue.JobDone:
		return "border-l-2 border-neon-green"
	case queue.JobFailed:
		return "border-l-2 border-neon-red bg-neon-red/5"
	default:
		return "border-l-2 border-border-dim"
	}
}

func jobDotClass(status queue.JobStatus) string {
	switch status {
	case queue.JobRunning:
		return "bg-neon-blue animate-pulse"
	case queue.JobDone:
		return "bg-neon-green"
	case queue.JobFailed:
		return "bg-neon-red"
	default:
		return "bg-dim"
	}
}

func jobFilters(counts any) []jobFilter {
	all, pending, running, done, failed := fieldInt(
		counts,
		"All",
	), fieldInt(
		counts,
		"Pending",
	), fieldInt(
		counts,
		"Running",
	), fieldInt(
		counts,
		"Done",
	), fieldInt(
		counts,
		"Failed",
	)
	active := fieldString(counts, "Active")
	if active == "" {
		active = "all"
	}

	return []jobFilter{
		{Status: "all", Label: "All", Count: all, Active: active == "all"},
		{Status: "pending", Label: "Pending", Count: pending, Active: active == "pending"},
		{Status: "running", Label: "Running", Count: running, Active: active == "running"},
		{Status: "done", Label: "Done", Count: done, Active: active == "done"},
		{Status: "failed", Label: "Failed", Count: failed, Active: active == "failed"},
	}
}

func fieldInt(value any, name string) int {
	v := reflectValue(value)
	if !v.IsValid() {
		return 0
	}
	field := v.FieldByName(name)
	if !field.IsValid() {
		return 0
	}
	return int(field.Int())
}

func fieldString(value any, name string) string {
	v := reflectValue(value)
	if !v.IsValid() {
		return ""
	}
	field := v.FieldByName(name)
	if !field.IsValid() {
		return ""
	}
	return field.String()
}

func reflectValue(value any) reflect.Value {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	return v
}

func times(count int) []int {
	if count < 0 {
		count = 0
	}
	values := make([]int, count)
	for i := range values {
		values[i] = i
	}
	return values
}

func randomHeight() int {
	return 180 + rand.Intn(180)
}

func New(svc *service.Service) *Server {
	funcMap := template.FuncMap{
		"actionSections": actionSections,
		"chapterRowData": chapterRowData,
		"jobStatusClass": jobStatusClass,
		"jobDotClass":    jobDotClass,
		"jobFilters":     jobFilters,
		"times":          times,
		"randomHeight":   randomHeight,
		"sub":            func(a, b int) int { return a - b },
		"add":            func(a, b int) int { return a + b },
		"countDownloaded": func(chapters []models.ChapterRecord) int {
			n := 0
			for _, c := range chapters {
				if c.Downloaded {
					n++
				}
			}
			return n
		},
	}
	templatesCache := make(map[string]*template.Template)

	patterns := []string{
		filepath.Join("web", "templates", "partials", "*.html"),
		filepath.Join("web", "templates", "components", "*.html"),
		filepath.Join("web", "templates", "layouts", "*.html"),
	}

	sharedFiles := []string{}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			slog.Error("server: global glob error", "pattern", pattern, "err", err)
			continue
		}
		sharedFiles = append(sharedFiles, matches...)
	}

	pages, err := filepath.Glob(filepath.Join("web", "templates", "pages", "*.html"))
	if err != nil {
		slog.Error("server: pages glob error", "err", err)
	}

	for _, page := range pages {
		name := filepath.Base(page)

		filesToParse := append([]string{page}, sharedFiles...)
		t, err := template.New(name).Funcs(funcMap).ParseFiles(filesToParse...)
		if err != nil {
			slog.Error("server: failed to parse template bundle", "page", name, "err", err)
			continue
		}
		templatesCache[name] = t
	}

	return &Server{
		service:   svc,
		templates: templatesCache,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	muxMu.Lock()
	if registeredMux == nil {
		registeredMux = routes(s, sessionSecret)
	}
	mux := registeredMux
	muxMu.Unlock()

	mux.ServeHTTP(w, r)
}
