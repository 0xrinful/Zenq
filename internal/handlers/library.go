package handlers

import (
	"html/template"
	"net/http"

	"github.com/0xrinful/Zenq/internal/service"
)

type Library struct {
	svc  *service.Service
	tmpl *template.Template
}

func NewLibrary(svc *service.Service, tmpl *template.Template) *Library {
	return &Library{svc: svc, tmpl: tmpl}
}

func (l *Library) Index(w http.ResponseWriter, r *http.Request) {
	writeTodo(w, r)
}
