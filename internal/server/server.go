package server

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/0xrinful/Zenq/internal/models"
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

func New(svc *service.Service) *Server {
	funcMap := template.FuncMap{
		"actionSections": actionSections,
		"chapterRowData": chapterRowData,
	}
	tmpl := template.New("root").Funcs(funcMap)
	patterns := []string{
		filepath.Join("web", "templates", "*"),
		filepath.Join("web", "templates", "*", "*"),
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
