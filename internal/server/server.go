package server

import (
	"html/template"
	"log/slog"
	"net/http"
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
	tmpl, err := template.ParseGlob("web/templates/**/*")
	if err != nil {
		slog.Warn("server: parse templates", "err", err)
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
