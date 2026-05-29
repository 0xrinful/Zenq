package server

import (
	"net/http"

	"github.com/0xrinful/Zenq/internal/handlers"
)

func routes(s *Server, secret string) *http.ServeMux {
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	auth := handlers.NewAuth(s.service, s.tmpl, func(w http.ResponseWriter, r *http.Request, userID int64) {
		SetSession(w, r, &Session{UserID: userID}, secret)
	}, func(w http.ResponseWriter) {
		ClearSession(w)
	})

	library := handlers.NewLibrary(s.service, s.tmpl)
	manga := handlers.NewManga(s.service, s.tmpl)
	chapter := handlers.NewChapter(s.service, s.tmpl)
	viewer := handlers.NewViewer(s.service, s.tmpl)
	sources := handlers.NewSources(s.service, s.tmpl)
	dashboard := handlers.NewDashboard(s.service, s.tmpl)

	mux.Handle("GET /login", http.HandlerFunc(auth.LoginPage))
	mux.Handle("POST /login", http.HandlerFunc(auth.LoginSubmit))
	mux.Handle("GET /signup", http.HandlerFunc(auth.SignupPage))
	mux.Handle("POST /signup", http.HandlerFunc(auth.SignupSubmit))
	mux.Handle("POST /logout", http.HandlerFunc(auth.Logout))

	requireAuth := func(h http.HandlerFunc) http.Handler {
		return AuthRequired(secret, h)
	}

	mux.Handle("GET /", requireAuth(library.Index))
	mux.Handle("GET /manga/{sourceID}/{slug}", requireAuth(manga.Detail))
	mux.Handle("POST /manga/{sourceID}/{slug}/download", requireAuth(manga.Download))
	mux.Handle("POST /manga/{sourceID}/{slug}/optimize", requireAuth(manga.Optimize))
	mux.Handle("POST /manga/{sourceID}/{slug}/pack", requireAuth(manga.Pack))
	mux.Handle("POST /manga/{sourceID}/{slug}/refresh", requireAuth(manga.Refresh))
	mux.Handle("DELETE /manga/{sourceID}/{slug}/files", requireAuth(manga.DeleteFiles))

	mux.Handle("POST /manga/{sourceID}/{slug}/ch/{num}/download", requireAuth(chapter.Download))
	mux.Handle("POST /manga/{sourceID}/{slug}/ch/{num}/optimize", requireAuth(chapter.Optimize))
	mux.Handle("POST /manga/{sourceID}/{slug}/ch/{num}/pack", requireAuth(chapter.Pack))
	mux.Handle("POST /manga/{sourceID}/{slug}/ch/{num}/read", requireAuth(chapter.ToggleRead))

	mux.Handle("GET /manga/{sourceID}/{slug}/ch/{num}", requireAuth(viewer.Page))
	mux.Handle("GET /manga/{sourceID}/{slug}/ch/{num}/pages", requireAuth(viewer.Pages))
	mux.Handle("GET /manga/{sourceID}/{slug}/ch/{num}/img/{file}", requireAuth(viewer.Image))

	mux.Handle("GET /sources", requireAuth(sources.Index))
	mux.Handle("GET /sources/{id}", requireAuth(sources.Browse))
	mux.Handle("GET /sources/{id}/search", requireAuth(sources.Search))
	mux.Handle("GET /sources/{id}/manga/{slug}", requireAuth(sources.MangaDetail))
	mux.Handle("POST /sources/{id}/manga/{slug}/import", requireAuth(sources.Import))

	mux.Handle("GET /dashboard", requireAuth(dashboard.Page))
	mux.Handle("GET /api/jobs", requireAuth(dashboard.Jobs))
	mux.Handle("GET /api/jobs/{id}", requireAuth(dashboard.JobDetail))
	mux.Handle("GET /api/storage", requireAuth(dashboard.Storage))
	mux.Handle("POST /api/flaresolver/start", requireAuth(dashboard.StartFlareSolver))

	return mux
}
