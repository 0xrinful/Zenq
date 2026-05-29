package server

import (
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/0xrinful/Zenq/internal/service"
)

var (
	muxMu         sync.Mutex
	registeredMux http.Handler
	sessionSecret string
)

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

func New(svc *service.Service) *Server {
	tmpl := template.New("root")
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
		tmpl = template.New("root")
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
