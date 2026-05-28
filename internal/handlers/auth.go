package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/0xrinful/Zenq/internal/service"
)

type Auth struct {
	svc          *service.Service
	tmpl         *template.Template
	setSession   func(http.ResponseWriter, *http.Request, int64)
	clearSession func(http.ResponseWriter)
}

type authPageData struct {
	Error string
}

func NewAuth(
	svc *service.Service,
	tmpl *template.Template,
	setSession func(http.ResponseWriter, *http.Request, int64),
	clearSession func(http.ResponseWriter),
) *Auth {
	return &Auth{
		svc:          svc,
		tmpl:         tmpl,
		setSession:   setSession,
		clearSession: clearSession,
	}
}

func (a *Auth) LoginPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if err := a.tmpl.ExecuteTemplate(w, "login.html", authPageData{}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *Auth) LoginSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := a.svc.SignIn(r.Context(), email, password)
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		_ = a.tmpl.ExecuteTemplate(w, "login.html", authPageData{Error: "Invalid email or password"})
		return
	}

	a.setSession(w, r, int64(user.ID))
	http.Redirect(w, r, "/", http.StatusFound)
}

func (a *Auth) SignupPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if err := a.tmpl.ExecuteTemplate(w, "signup.html", authPageData{}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *Auth) SignupSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := a.svc.SignUp(r.Context(), email, password)
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		message := "Unable to create account"
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.email") {
			message = "Email already exists"
		}
		_ = a.tmpl.ExecuteTemplate(w, "signup.html", authPageData{Error: message})
		return
	}

	a.setSession(w, r, int64(user.ID))
	http.Redirect(w, r, "/", http.StatusFound)
}

func (a *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	a.clearSession(w)
	http.Redirect(w, r, "/login", http.StatusFound)
}

func writeTodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "TODO: %s", template.HTMLEscapeString(r.URL.Path))
}
