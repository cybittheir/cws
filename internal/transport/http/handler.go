package http

import (
	"corporate-workspace/internal/config"
	"corporate-workspace/internal/domain/directory"
	"corporate-workspace/internal/domain/identity"
	"corporate-workspace/internal/repository"
	"corporate-workspace/internal/service"
	"embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strconv"
)

//go:embed web/templates/*.html web/static/*
var assets embed.FS

type Handler struct {
	cfg       config.Config
	logger    *log.Logger
	templates *template.Template
	store     repository.Store
	directory *service.DirectoryService
}
type Page struct {
	Title     string
	User      identity.User
	Companies []identity.Membership
	Company   identity.Membership
	Result    repository.DirectoryResult
	ShowCity  bool
	Employee  *directory.Employee
	Error     string
	Success   string
}

func NewHandler(cfg config.Config, logger *log.Logger, store repository.Store, d *service.DirectoryService) (http.Handler, error) {
	t, err := template.ParseFS(assets, "web/templates/*.html")
	if err != nil {
		return nil, err
	}
	h := &Handler{cfg, logger, t, store, d}
	m := http.NewServeMux()
	m.HandleFunc("GET /", h.home)
	m.HandleFunc("GET /login", h.loginPage)
	m.HandleFunc("POST /login", h.login)
	m.HandleFunc("POST /logout", h.logout)
	m.HandleFunc("GET /directory", h.auth(h.directoryPage))
	m.HandleFunc("GET /profile", h.auth(h.profile))
	m.HandleFunc("GET /profile/edit", h.auth(h.editProfilePage))
	m.HandleFunc("POST /profile/edit", h.auth(h.updateProfile))
	m.HandleFunc("GET /switch-company", h.auth(h.switchCompany))
	m.HandleFunc("GET /healthz", h.health)
	st, err := fs.Sub(assets, "web/static")
	if err != nil {
		return nil, err
	}
	m.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(st))))
	return h.logging(m), nil
}
func (h *Handler) home(w http.ResponseWriter, r *http.Request) {
	if _, e := r.Cookie("cw_session"); e == nil {
		http.Redirect(w, r, "/directory", 302)
		return
	}
	http.Redirect(w, r, "/login", 302)
}
func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "version": h.cfg.App.Version})
}
func (h *Handler) loginPage(w http.ResponseWriter, r *http.Request) {
	h.render(w, "login.html", Page{Title: "Вход"})
}
func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	u, e := h.store.Authenticate(r.Context(), r.FormValue("email"), r.FormValue("password"))
	if e != nil {
		h.renderStatus(w, 401, "login.html", Page{Title: "Вход", Error: "Неверный email или пароль. Используйте demo / demo."})
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "cw_session", Value: strconv.Itoa(u.ID), Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: 28800, Secure: h.cfg.Security.CookieSecure})
	http.Redirect(w, r, "/directory", 302)
}
func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "cw_session", Value: "", Path: "/", MaxAge: -1, HttpOnly: true})
	http.Redirect(w, r, "/login", 302)
}
func (h *Handler) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, e := r.Cookie("cw_session")
		if e != nil {
			http.Redirect(w, r, "/login", 302)
			return
		}
		id, e := strconv.Atoi(c.Value)
		if e != nil {
			http.Redirect(w, r, "/login", 302)
			return
		}
		u, e := h.store.UserByID(r.Context(), id)
		if e != nil {
			http.Redirect(w, r, "/login", 302)
			return
		}
		next.ServeHTTP(w, r.WithContext(contextWithUser(r.Context(), u)))
	}
}
