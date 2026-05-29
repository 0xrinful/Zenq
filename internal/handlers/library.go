package handlers

import (
	"html/template"
	"net/http"

	"github.com/0xrinful/Zenq/internal/models"
	"github.com/0xrinful/Zenq/internal/service"
)

type Library struct {
	svc       *service.Service
	templates map[string]*template.Template
}

type libraryData struct {
	CurrentPath string
	Mangas      []models.MangaRecord
}

func NewLibrary(svc *service.Service, templates map[string]*template.Template) *Library {
	return &Library{svc: svc, templates: templates}
}

func (l *Library) Index(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	mangas, err := l.svc.Library(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderTemplate(w, l.templates, "library.html", libraryData{
		CurrentPath: "library",
		Mangas:      mangas,
	})
}
