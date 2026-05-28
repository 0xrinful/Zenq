package handlers

import (
	"html/template"
	"net/http"

	"github.com/0xrinful/Zenq/internal/service"
)

type Viewer struct {
	svc  *service.Service
	tmpl *template.Template
}

func NewViewer(svc *service.Service, tmpl *template.Template) *Viewer {
	return &Viewer{svc: svc, tmpl: tmpl}
}

func (v *Viewer) Page(w http.ResponseWriter, r *http.Request) {
	writeTodo(w, r)
}

func (v *Viewer) Pages(w http.ResponseWriter, r *http.Request) {
	writeTodo(w, r)
}
