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
	Label    string
	Endpoint string
	HasRange bool
	IsDelete bool
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
	service *service.Service
	tmpl    *template.Template
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
	all, pending, running, done, failed := fieldInt(counts, "All"), fieldInt(counts, "Pending"), fieldInt(counts, "Running"), fieldInt(counts, "Done"), fieldInt(counts, "Failed")
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
	}
	tmpl := template.New("root").Funcs(funcMap)
	patterns := []string{
		filepath.Join("web", "templates", "*.html"),
		filepath.Join("web", "templates", "*", "*.html"),
	}

	parsed := false
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			slog.Warn("server: glob templates", "pattern", pattern, "err", err)
			continue
		}
		if len(matches) == 0 {
			continue
		}
		if _, err := tmpl.ParseGlob(pattern); err != nil {
			slog.Warn("server: parse templates", "pattern", pattern, "err", err)
			continue
		}
		parsed = true
	}
	if !parsed {
		tmpl = template.New("root").Funcs(funcMap)
	}

	return &Server{
		service: svc,
		tmpl:    tmpl,
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
