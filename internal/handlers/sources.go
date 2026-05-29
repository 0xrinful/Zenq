package handlers

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/0xrinful/Zenq/internal/models"
	"github.com/0xrinful/Zenq/internal/service"
	"github.com/0xrinful/Zenq/internal/sources"
)

type Sources struct {
	svc       *service.Service
	templates map[string]*template.Template
}

type sourcesPageData struct {
	CurrentPath  string
	PageTitle    string
	BodyTemplate string
}

type sourcesListData struct {
	sourcesPageData
	Sources []sources.SourceInfo
}

type sourceBrowseData struct {
	sourcesPageData
	SourceID   string
	SourceName string
}

type browseResultsData struct {
	SourceID string
	Mangas   []models.Manga
	Page     int
}

type sourceMangaData struct {
	sourcesPageData
	SourceID  string
	Manga     models.Manga
	InLibrary bool
}

func NewSources(svc *service.Service, templates map[string]*template.Template) *Sources {
	return &Sources{svc: svc, templates: templates}
}

func (s *Sources) Index(w http.ResponseWriter, r *http.Request) {
	infos := s.svc.Sources()
	renderTemplateName(w, s.templates, "sources.html", "sources-list", sourcesListData{
		sourcesPageData: sourcesPageData{
			CurrentPath:  "sources",
			PageTitle:    "Sources — ZENQ",
			BodyTemplate: "sources-list-body",
		},
		Sources: infos,
	})
}

func (s *Sources) Browse(w http.ResponseWriter, r *http.Request) {
	sourceID := r.PathValue("id")
	info, ok := s.svc.SourceInfo(sourceID)
	if !ok {
		http.NotFound(w, r)
		return
	}

	renderTemplateName(w, s.templates, "sources.html", "source-browse", sourceBrowseData{
		sourcesPageData: sourcesPageData{
			CurrentPath:  "sources",
			PageTitle:    info.Name,
			BodyTemplate: "source-browse-body",
		},
		SourceID:   sourceID,
		SourceName: info.Name,
	})
}

func (s *Sources) BrowseResults(w http.ResponseWriter, r *http.Request) {
	sourceID := r.PathValue("id")

	pageStr := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	size := 25

	mangas, err := s.svc.SourceLatest(r.Context(), sourceID, page, size)
	if err != nil {
		if errors.Is(err, service.ErrUnknownSource) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderTemplateName(w, s.templates, "sources.html", "browse-results-partial", browseResultsData{
		SourceID: sourceID,
		Mangas:   mangas,
		Page:     page,
	})
}

func (s *Sources) Search(w http.ResponseWriter, r *http.Request) {
	sourceID := r.PathValue("id")
	query := r.URL.Query().Get("q")
	mangas, err := s.svc.SourceSearch(r.Context(), sourceID, query)
	if err != nil {
		if errors.Is(err, service.ErrUnknownSource) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderTemplateName(w, s.templates, "sources.html", "browse-results-partial", browseResultsData{
		SourceID: sourceID,
		Mangas:   mangas,
	})
}

func (s *Sources) MangaDetail(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	sourceID := r.PathValue("id")
	slug := r.PathValue("slug")

	manga, err := s.svc.SourceManga(r.Context(), sourceID, slug)
	if err != nil {
		if errors.Is(err, service.ErrUnknownSource) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if manga == nil {
		http.NotFound(w, r)
		return
	}

	library, err := s.svc.Library(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	inLibrary := false
	for _, entry := range library {
		if entry.Slug == slug && entry.SourceID == sourceID {
			inLibrary = true
			break
		}
	}

	renderTemplateName(w, s.templates, "sources.html", "source-manga-detail", sourceMangaData{
		sourcesPageData: sourcesPageData{
			CurrentPath:  "sources",
			PageTitle:    fmt.Sprintf("%s — ZENQ", manga.Title),
			BodyTemplate: "source-manga-detail-body",
		},
		SourceID:  sourceID,
		Manga:     *manga,
		InLibrary: inLibrary,
	})
}

func (s *Sources) Import(w http.ResponseWriter, r *http.Request) {
	sourceID := r.PathValue("id")
	slug := r.PathValue("slug")

	if err := s.svc.ImportManga(r.Context(), sourceID, slug); err != nil {
		writeActionError(w, err)
		return
	}

	writeToast(w, "Imported", "success")
	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte(fmt.Sprintf(`<div class="flex items-center gap-3">
  <span class="text-neon-green text-xs font-mono">✓ In Library</span>
  <a href="/manga/%s/%s" class="btn-ghost text-xs">View →</a>
</div>`, sourceID, slug)))
}
