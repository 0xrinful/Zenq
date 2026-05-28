package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/0xrinful/Zenq/internal/models"
	"github.com/0xrinful/Zenq/internal/service"
)

type Manga struct {
	svc  *service.Service
	tmpl *template.Template
}

type RangeRequest struct {
	From  float64 `json:"from"`
	To    float64 `json:"to"`
	All   bool    `json:"all"`
	Force bool    `json:"force"`
}

func NewManga(svc *service.Service, tmpl *template.Template) *Manga {
	return &Manga{svc: svc, tmpl: tmpl}
}

func (m *Manga) Detail(w http.ResponseWriter, r *http.Request) {
	writeTodo(w, r)
}

func (m *Manga) Download(w http.ResponseWriter, r *http.Request) {
	var req RangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rangeReq := models.ChapterRange{From: req.From, To: req.To, All: req.All, Force: req.Force}
	sourceID := r.PathValue("sourceID")
	slug := r.PathValue("slug")
	if _, err := m.svc.DownloadRange(r.Context(), sourceID, slug, rangeReq); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeTodo(w, r)
}

func (m *Manga) Optimize(w http.ResponseWriter, r *http.Request) {
	var req RangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rangeReq := models.ChapterRange{From: req.From, To: req.To, All: req.All, Force: req.Force}
	sourceID := r.PathValue("sourceID")
	slug := r.PathValue("slug")
	if _, err := m.svc.OptimizeRange(r.Context(), sourceID, slug, rangeReq); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeTodo(w, r)
}

func (m *Manga) Pack(w http.ResponseWriter, r *http.Request) {
	var req RangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rangeReq := models.ChapterRange{From: req.From, To: req.To, All: req.All, Force: req.Force}
	sourceID := r.PathValue("sourceID")
	slug := r.PathValue("slug")
	if _, err := m.svc.PackRange(r.Context(), sourceID, slug, rangeReq); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeTodo(w, r)
}

func (m *Manga) Refresh(w http.ResponseWriter, r *http.Request) {
	sourceID := r.PathValue("sourceID")
	slug := r.PathValue("slug")
	if err := m.svc.RefreshManga(sourceID, slug); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeTodo(w, r)
}

func (m *Manga) DeleteFiles(w http.ResponseWriter, r *http.Request) {
	var req service.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sourceID := r.PathValue("sourceID")
	slug := r.PathValue("slug")
	if err := m.svc.DeleteMangaFiles(sourceID, slug, req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeTodo(w, r)
}
