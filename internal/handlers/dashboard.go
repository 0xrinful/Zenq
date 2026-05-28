package handlers

import (
	"html/template"
	"net/http"

	"github.com/0xrinful/Zenq/internal/service"
)

type Dashboard struct {
	svc  *service.Service
	tmpl *template.Template
}

func NewDashboard(svc *service.Service, tmpl *template.Template) *Dashboard {
	return &Dashboard{svc: svc, tmpl: tmpl}
}

func (d *Dashboard) Page(w http.ResponseWriter, r *http.Request) {
	writeTodo(w, r)
}

func (d *Dashboard) Jobs(w http.ResponseWriter, r *http.Request) {
	writeTodo(w, r)
}

func (d *Dashboard) JobDetail(w http.ResponseWriter, r *http.Request) {
	writeTodo(w, r)
}

func (d *Dashboard) Storage(w http.ResponseWriter, r *http.Request) {
	writeTodo(w, r)
}

func (d *Dashboard) StartFlareSolver(w http.ResponseWriter, r *http.Request) {
	writeTodo(w, r)
}
