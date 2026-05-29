package handlers

import (
	"errors"
	"html/template"
	"net/http"
	"strconv"

	"github.com/0xrinful/Zenq/internal/models"
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
	noSwap(w)

	chapter, _, err := c.lookupChapter(r)
	if err != nil {
		writeActionError(w, err)
		return
	}

	sourceID := r.PathValue("sourceID")
	if _, err := c.svc.DownloadChapter(r.Context(), sourceID, chapter.Chapter); err != nil {
		writeActionError(w, err)
		return
	}

	writeToast(w, "Download queued", "success")
	w.WriteHeader(http.StatusOK)
}

func (c *Chapter) Optimize(w http.ResponseWriter, r *http.Request) {
	noSwap(w)

	chapter, _, err := c.lookupChapter(r)
	if err != nil {
		writeActionError(w, err)
		return
	}

	if _, err := c.svc.OptimizeChapter(r.Context(), chapter.Chapter); err != nil {
		writeActionError(w, err)
		return
	}

	writeToast(w, "Optimize queued", "success")
	w.WriteHeader(http.StatusOK)
}

func (c *Chapter) Pack(w http.ResponseWriter, r *http.Request) {
	noSwap(w)

	chapter, _, err := c.lookupChapter(r)
	if err != nil {
		writeActionError(w, err)
		return
	}

	if _, err := c.svc.PackChapter(r.Context(), chapter.Chapter); err != nil {
		writeActionError(w, err)
		return
	}

	writeToast(w, "Pack queued", "success")
	w.WriteHeader(http.StatusOK)
}

func (c *Chapter) ToggleRead(w http.ResponseWriter, r *http.Request) {
	noSwap(w)

	chapter, readMarks, err := c.lookupChapter(r)
	if err != nil {
		writeActionError(w, err)
		return
	}

	userID := getUserID(r.Context())
	sourceID := r.PathValue("sourceID")
	slug := r.PathValue("slug")
	number := chapter.Number

	message := "Marked read"
	if readMarks[number] {
		if err := c.svc.MarkUnread(r.Context(), userID, slug, sourceID, number); err != nil {
			writeActionError(w, err)
			return
		}
		message = "Marked unread"
	} else {
		if err := c.svc.MarkRead(r.Context(), userID, slug, sourceID, number); err != nil {
			writeActionError(w, err)
			return
		}
	}

	writeToast(w, message, "success")
	w.WriteHeader(http.StatusOK)
}

func (c *Chapter) lookupChapter(r *http.Request) (*models.ChapterRecord, map[float64]bool, error) {
	userID := getUserID(r.Context())
	sourceID := r.PathValue("sourceID")
	slug := r.PathValue("slug")
	rawNum := r.PathValue("num")
	number, err := strconv.ParseFloat(rawNum, 64)
	if err != nil {
		return nil, nil, err
	}

	page, err := c.svc.MangaPage(r.Context(), userID, slug, sourceID)
	if err != nil {
		return nil, nil, err
	}

	readMarks := make(map[float64]bool, len(page.ReadMarks))
	for _, read := range page.ReadMarks {
		readMarks[read] = true
	}

	for i := range page.Chapters {
		if page.Chapters[i].Number == number {
			return &page.Chapters[i], readMarks, nil
		}
	}

	return nil, readMarks, errors.New("chapter not found")
}
