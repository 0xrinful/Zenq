package handlers

import (
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/0xrinful/Zenq/internal/models"
	"github.com/0xrinful/Zenq/internal/service"
)

type Manga struct {
	svc       *service.Service
	templates map[string]*template.Template
}

type RangeRequest struct {
	From  float64
	To    float64
	All   bool
	Force bool
}

func (req *RangeRequest) Parse(r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	fromStr := r.FormValue("from")
	toStr := r.FormValue("to")

	if fromStr != "" {
		from, err := strconv.ParseFloat(fromStr, 64)
		if err != nil {
			return fmt.Errorf("invalid from: %w", err)
		}
		req.From = from
	}

	if toStr != "" {
		to, err := strconv.ParseFloat(toStr, 64)
		if err != nil {
			return fmt.Errorf("invalid to: %w", err)
		}
		req.To = to
	}

	req.All = r.FormValue("all") == "on"
	req.Force = r.FormValue("force") == "on"
	return nil
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

func (m *Manga) Unfavorite(w http.ResponseWriter, r *http.Request) {
	sourceID := r.PathValue("sourceID")
	slug := r.PathValue("slug")
	userID := getUserID(r.Context())

	err := m.svc.Unfavorite(r.Context(), userID, sourceID, slug)
	if err != nil {
		writeActionError(w, err)
		return
	}

	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
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
	if err := req.Parse(r); err != nil {
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
	if err := req.Parse(r); err != nil {
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
	if err := req.Parse(r); err != nil {
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

func (m *Manga) DownloadPacked(w http.ResponseWriter, r *http.Request) {
	noSwap(w)

	var req RangeRequest
	if err := req.Parse(r); err != nil {
		writeActionError(w, err)
		return
	}

	rangeReq := models.ChapterRange{From: req.From, To: req.To, All: req.All, Force: req.Force}
	sourceID := r.PathValue("sourceID")
	slug := r.PathValue("slug")

	destZip, err := m.svc.PackManga(r.Context(), sourceID, slug, rangeReq)
	if err != nil {
		writeActionError(w, err)
		return
	}

	defer func() {
		if err := os.Remove(destZip); err != nil {
			slog.Error("failed to remove temp file", "name", destZip, "err", err)
		}
	}()

	w.Header().
		Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q.zip", slug))
	w.Header().Set("Content-Type", "application/zip")

	http.ServeFile(w, r, destZip)
}

func (m *Manga) DeleteFiles(w http.ResponseWriter, r *http.Request) {
	noSwap(w)

	if err := r.ParseForm(); err != nil {
		writeActionError(w, err)
		return
	}

	req := service.DeleteRequest{
		Raw:       r.FormValue("raw") == "on",
		Optimized: r.FormValue("optimized") == "on",
		Packed:    r.FormValue("packed") == "on",
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
