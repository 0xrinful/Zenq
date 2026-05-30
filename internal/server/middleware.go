package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/0xrinful/Zenq/internal/contextkeys"
)

func AuthRequired(secret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := GetSession(r, secret)
		if err != nil || session == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		ctx := context.WithValue(r.Context(), contextkeys.UserID, int(session.UserID))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AdminRequired(secret string, next http.Handler) http.Handler {
	return AuthRequired(secret, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(contextkeys.UserID).(int)
		if userID != 1 {
			w.Header().Set("X-Toast", `{"message":"this action require an admin","type":"error"}`)
			w.WriteHeader(http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}))
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(p)
}

func WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w}
		next.ServeHTTP(sw, r)
		duration := time.Since(start)

		status := sw.status
		if status == 0 {
			status = http.StatusOK
		}

		slog.Info(
			"request",
			"method",
			r.Method,
			"path",
			r.URL.Path,
			"status",
			status,
			"duration",
			duration,
		)
	})
}
