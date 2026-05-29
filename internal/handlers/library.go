package handlers

import (
	"html/template"
	"net/http"

	"github.com/0xrinful/Zenq/internal/models"
	"github.com/0xrinful/Zenq/internal/service"
)

type Library struct {
	svc  *service.Service
	tmpl *template.Template
}

type libraryData struct {
	CurrentPath string
	Mangas      []models.MangaRecord
}

func NewLibrary(svc *service.Service, tmpl *template.Template) *Library {
	return &Library{svc: svc, tmpl: tmpl}
}

func (l *Library) Index(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	mangas, err := l.svc.Library(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderTemplate(w, l.tmpl, "library.html", libraryData{
		CurrentPath: "library",
		Mangas:      mangas,
	})
}
