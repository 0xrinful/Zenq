package handlers

import (
	"context"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/0xrinful/Zenq/internal/contextkeys"
)

type toastMessage struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

func getUserID(ctx context.Context) int {
	if ctx == nil {
		return 0
	}
	if id, ok := ctx.Value(contextkeys.UserID).(int); ok {
		return id
	}
	if id, ok := ctx.Value(contextkeys.UserID).(int64); ok {
		return int(id)
	}
	return 0
}

func renderTemplate(
	w http.ResponseWriter,
	templates map[string]*template.Template,
	name string,
	data any,
) {
	w.Header().Set("Content-Type", "text/html")

	t, ok := templates[name]
	if !ok {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderTemplateName(
	w http.ResponseWriter,
	templates map[string]*template.Template,
	name string,
	tmpl string,
	data any,
) {
	w.Header().Set("Content-Type", "text/html")

	t, ok := templates[name]
	if !ok {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	if err := t.ExecuteTemplate(w, tmpl, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeToast(w http.ResponseWriter, message, tone string) {
	payload, err := json.Marshal(toastMessage{Message: message, Type: tone})
	if err != nil {
		return
	}
	w.Header().Set("X-Toast", string(payload))
}

func writeActionError(w http.ResponseWriter, err error) {
	writeToast(w, err.Error(), "error")
	w.WriteHeader(http.StatusUnprocessableEntity)
}

func noSwap(w http.ResponseWriter) {
	w.Header().Set("HX-Reswap", "none")
}
