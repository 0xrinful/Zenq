package handlers

import (
	"html/template"
	"net/http"

	"github.com/0xrinful/Zenq/internal/service"
)

type Chapter struct {
	svc  *service.Service
	tmpl *template.Template
}

func NewChapter(svc *service.Service, tmpl *template.Template) *Chapter {
	return &Chapter{svc: svc, tmpl: tmpl}
}

func (c *Chapter) Download(w http.ResponseWriter, r *http.Request) {
	writeTodo(w, r)
}

func (c *Chapter) Optimize(w http.ResponseWriter, r *http.Request) {
	writeTodo(w, r)
}

func (c *Chapter) Pack(w http.ResponseWriter, r *http.Request) {
	writeTodo(w, r)
}

func (c *Chapter) ToggleRead(w http.ResponseWriter, r *http.Request) {
	writeTodo(w, r)
}
