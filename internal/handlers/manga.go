package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/0xrinful/Zenq/internal/models"
	"github.com/0xrinful/Zenq/internal/service"
)

type Manga struct {
	svc       *service.Service
	templates map[string]*template.Template
}

type RangeRequest struct {
	From  float64 `json:"from"`
	To    float64 `json:"to"`
	All   bool    `json:"all"`
	Force bool    `json:"force"`
}

type mangaPageData struct {
	CurrentPath string
	SourceID    string
	MangaSlug   string
	Manga       models.MangaRecord
	Chapters    []models.ChapterRecord
	ReadMarks   map[float64]bool
}

func NewManga(svc *service.Service, templates map[string]*template.Template) *Manga {
	return &Manga{svc: svc, templates: templates}
}

func (m *Manga) Detail(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	sourceID := r.PathValue("sourceID")
	slug := r.PathValue("slug")

	result, err := m.svc.MangaPage(r.Context(), userID, slug, sourceID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	chapters := append([]models.ChapterRecord(nil), result.Chapters...)
	sort.Slice(chapters, func(i, j int) bool {
		return chapters[i].Number > chapters[j].Number
	})

	readMarks := make(map[float64]bool, len(result.ReadMarks))
	for _, number := range result.ReadMarks {
		readMarks[number] = true
	}

	renderTemplate(w, m.templates, "manga.html", mangaPageData{
		CurrentPath: "library",
		SourceID:    sourceID,
		MangaSlug:   slug,
		Manga:       *result.Manga,
		Chapters:    chapters,
		ReadMarks:   readMarks,
	})
}

func (m *Manga) Cover(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	sourceID := r.PathValue("sourceID")
	slug := r.PathValue("slug")

	result, err := m.svc.MangaPage(r.Context(), userID, slug, sourceID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if result.Manga == nil || result.Manga.CoverPath == "" {
		http.NotFound(w, r)
		return
	}

	coverPath, err := m.svc.Files().ResolvePath(result.Manga.CoverPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	file, err := os.Open(coverPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if info.IsDir() {
		http.NotFound(w, r)
		return
	}

	switch strings.ToLower(filepath.Ext(coverPath)) {
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	default:
		w.Header().Set("Content-Type", "image/jpeg")
	}
	http.ServeContent(w, r, info.Name(), info.ModTime(), file)
}

func (m *Manga) Download(w http.ResponseWriter, r *http.Request) {
	noSwap(w)

	var req RangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeActionError(w, err)
		return
	}

	rangeReq := models.ChapterRange{From: req.From, To: req.To, All: req.All, Force: req.Force}
	sourceID := r.PathValue("sourceID")
	slug := r.PathValue("slug")
	jobIDs, err := m.svc.DownloadRange(r.Context(), sourceID, slug, rangeReq)
	if err != nil {
		writeActionError(w, err)
		return
	}

	writeToast(w, fmt.Sprintf("Download queued (%d jobs)", len(jobIDs)), "success")
	w.WriteHeader(http.StatusOK)
}

func (m *Manga) Optimize(w http.ResponseWriter, r *http.Request) {
	noSwap(w)

	var req RangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeActionError(w, err)
		return
	}

	rangeReq := models.ChapterRange{From: req.From, To: req.To, All: req.All, Force: req.Force}
	sourceID := r.PathValue("sourceID")
	slug := r.PathValue("slug")
	jobIDs, err := m.svc.OptimizeRange(r.Context(), sourceID, slug, rangeReq)
	if err != nil {
		writeActionError(w, err)
		return
	}

	writeToast(w, fmt.Sprintf("Optimize queued (%d jobs)", len(jobIDs)), "success")
	w.WriteHeader(http.StatusOK)
}

func (m *Manga) Pack(w http.ResponseWriter, r *http.Request) {
	noSwap(w)

	var req RangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeActionError(w, err)
		return
	}

	rangeReq := models.ChapterRange{From: req.From, To: req.To, All: req.All, Force: req.Force}
	sourceID := r.PathValue("sourceID")
	slug := r.PathValue("slug")
	jobIDs, err := m.svc.PackRange(r.Context(), sourceID, slug, rangeReq)
	if err != nil {
		writeActionError(w, err)
		return
	}

	writeToast(w, fmt.Sprintf("Pack queued (%d jobs)", len(jobIDs)), "success")
	w.WriteHeader(http.StatusOK)
}

func (m *Manga) Refresh(w http.ResponseWriter, r *http.Request) {
	noSwap(w)

	sourceID := r.PathValue("sourceID")
	slug := r.PathValue("slug")
	if err := m.svc.RefreshManga(sourceID, slug); err != nil {
		writeActionError(w, err)
		return
	}

	writeToast(w, "Sync started", "success")
	w.WriteHeader(http.StatusOK)
}

func (m *Manga) DeleteFiles(w http.ResponseWriter, r *http.Request) {
	noSwap(w)

	var req service.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeActionError(w, err)
		return
	}

	sourceID := r.PathValue("sourceID")
	slug := r.PathValue("slug")
	if err := m.svc.DeleteMangaFiles(sourceID, slug, req); err != nil {
		writeActionError(w, err)
		return
	}

	writeToast(w, "Files deleted", "success")
	w.WriteHeader(http.StatusOK)
}
