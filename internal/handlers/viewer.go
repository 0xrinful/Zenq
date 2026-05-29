package handlers

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"

	"github.com/0xrinful/Zenq/internal/models"
	"github.com/0xrinful/Zenq/internal/service"
)

type Viewer struct {
	svc       *service.Service
	templates map[string]*template.Template
}

type PageResponse struct {
	URLs []string `json:"urls"`
}

type viewerData struct {
	SourceID      string
	MangaSlug     string
	MangaTitle    string
	ChapterNum    float64
	PrevChapter   float64
	NextChapter   float64
	TotalChapters int
	ChapterList   []models.ChapterRecord
}

var errChapterNotFound = errors.New("chapter not found")

func NewViewer(svc *service.Service, templates map[string]*template.Template) *Viewer {
	return &Viewer{svc: svc, templates: templates}
}

func (v *Viewer) Page(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	sourceID := r.PathValue("sourceID")
	slug := r.PathValue("slug")
	rawNum := r.PathValue("num")
	number, err := strconv.ParseFloat(rawNum, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := v.svc.MangaPage(r.Context(), userID, slug, sourceID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !chapterExists(result.Chapters, number) {
		http.NotFound(w, r)
		return
	}

	chaptersAsc := append([]models.ChapterRecord(nil), result.Chapters...)
	sort.Slice(chaptersAsc, func(i, j int) bool {
		return chaptersAsc[i].Number < chaptersAsc[j].Number
	})

	prev, next := findPrevNext(chaptersAsc, number)

	chapterList := append([]models.ChapterRecord(nil), result.Chapters...)
	sort.Slice(chapterList, func(i, j int) bool {
		return chapterList[i].Number > chapterList[j].Number
	})

	if err := v.svc.MarkRead(r.Context(), userID, slug, sourceID, number); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderTemplate(w, v.templates, "viewer.html", viewerData{
		SourceID:      sourceID,
		MangaSlug:     slug,
		MangaTitle:    result.Manga.Title,
		ChapterNum:    number,
		PrevChapter:   prev,
		NextChapter:   next,
		TotalChapters: len(chapterList),
		ChapterList:   chapterList,
	})
}

func (v *Viewer) Pages(w http.ResponseWriter, r *http.Request) {
	_, chapter, _, err := v.lookupChapter(r)
	if err != nil {
		v.writePageError(w, r, err)
		return
	}

	sourceID := r.PathValue("sourceID")
	slug := r.PathValue("slug")
	rawNum := r.PathValue("num")

	dir, ok := chapterDir(chapter)
	if !ok {
		writeJSON(w, http.StatusNotFound, PageResponse{URLs: nil})
		return
	}

	absDir, err := v.svc.Files().ResolvePath(dir)
	if err != nil {
		writeJSON(w, http.StatusNotFound, PageResponse{URLs: nil})
		return
	}

	entries, err := os.ReadDir(absDir)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSON(w, http.StatusNotFound, PageResponse{URLs: nil})
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	urls := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		escapedName := url.PathEscape(entry.Name())
		urls = append(urls, "/manga/"+sourceID+"/"+slug+"/ch/"+rawNum+"/img/"+escapedName)
	}

	writeJSON(w, http.StatusOK, PageResponse{URLs: urls})
}

func (v *Viewer) Image(w http.ResponseWriter, r *http.Request) {
	_, chapter, _, err := v.lookupChapter(r)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) || errors.Is(err, errChapterNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dir, ok := chapterDir(chapter)
	if !ok {
		http.NotFound(w, r)
		return
	}

	fileName := r.PathValue("file")
	filePath, err := v.svc.Files().ResolveFile(dir, fileName)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	file, err := os.Open(filePath)
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

	http.ServeContent(w, r, info.Name(), info.ModTime(), file)
}

func (v *Viewer) lookupChapter(
	r *http.Request,
) (*service.MangaPageResult, *models.ChapterRecord, float64, error) {
	userID := getUserID(r.Context())
	sourceID := r.PathValue("sourceID")
	slug := r.PathValue("slug")
	rawNum := r.PathValue("num")
	number, err := strconv.ParseFloat(rawNum, 64)
	if err != nil {
		return nil, nil, 0, err
	}

	page, err := v.svc.MangaPage(r.Context(), userID, slug, sourceID)
	if err != nil {
		return nil, nil, number, err
	}

	for i := range page.Chapters {
		if page.Chapters[i].Number == number {
			return page, &page.Chapters[i], number, nil
		}
	}

	return page, nil, number, errChapterNotFound
}

func (v *Viewer) writePageError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, service.ErrNotFound) || errors.Is(err, errChapterNotFound) {
		writeJSON(w, http.StatusNotFound, PageResponse{URLs: nil})
		return
	}
	var numErr *strconv.NumError
	if errors.As(err, &numErr) {
		writeJSON(w, http.StatusBadRequest, PageResponse{URLs: nil})
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(data)
}

func chapterDir(chapter *models.ChapterRecord) (string, bool) {
	if chapter == nil {
		return "", false
	}
	if chapter.Optimized && chapter.OptimizedPath != "" {
		return chapter.OptimizedPath, true
	}
	if chapter.Downloaded && chapter.RawPath != "" {
		return chapter.RawPath, true
	}
	return "", false
}

func chapterExists(chapters []models.ChapterRecord, number float64) bool {
	for _, ch := range chapters {
		if ch.Number == number {
			return true
		}
	}
	return false
}

func findPrevNext(chapters []models.ChapterRecord, number float64) (float64, float64) {
	prev := 0.0
	next := 0.0
	for i, ch := range chapters {
		if ch.Number != number {
			continue
		}
		if i > 0 {
			prev = chapters[i-1].Number
		}
		if i+1 < len(chapters) {
			next = chapters[i+1].Number
		}
		break
	}
	return prev, next
}
