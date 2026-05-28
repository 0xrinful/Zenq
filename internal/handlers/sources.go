package handlers

import (
	"html/template"
	"net/http"

	"github.com/0xrinful/Zenq/internal/service"
)

type Sources struct {
	svc  *service.Service
	tmpl *template.Template
}

func NewSources(svc *service.Service, tmpl *template.Template) *Sources {
	return &Sources{svc: svc, tmpl: tmpl}
}

func (s *Sources) Index(w http.ResponseWriter, r *http.Request) {
	writeTodo(w, r)
}

func (s *Sources) Browse(w http.ResponseWriter, r *http.Request) {
	writeTodo(w, r)
}

func (s *Sources) Search(w http.ResponseWriter, r *http.Request) {
	writeTodo(w, r)
}

func (s *Sources) MangaDetail(w http.ResponseWriter, r *http.Request) {
	writeTodo(w, r)
}

func (s *Sources) Import(w http.ResponseWriter, r *http.Request) {
	writeTodo(w, r)
}
