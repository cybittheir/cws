package http

import (
	"context"
	"corporate-workspace/internal/config"
	"corporate-workspace/internal/domain"
	"corporate-workspace/internal/repository"
	sqlstore "corporate-workspace/internal/repository/sqlite"
	"corporate-workspace/internal/service"
	"embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"time"
)

//go:embed web/templates/*.html web/static/*
var assets embed.FS

type Handler struct {
	cfg    config.Config
	logger *log.Logger
	t      *template.Template
	store  repository.Store
	dir    *service.DirectoryService
}
type Page struct {
	Title         string
	User          domain.User
	Employees     []domain.Employee
	Departments   []domain.Department
	Total, Cities int
	Employee      *domain.Employee
	Error         string
}
type userKey struct{}

func NewHandler(c config.Config, l *log.Logger, s repository.Store, d *service.DirectoryService) (http.Handler, error) {
	t, e := template.ParseFS(assets, "web/templates/*.html")
	if e != nil {
		return nil, e
	}
	h := &Handler{c, l, t, s, d}
	m := http.NewServeMux()
	m.HandleFunc("GET /", h.home)
	m.HandleFunc("GET /login", h.loginPage)
	m.HandleFunc("POST /login", h.login)
	m.HandleFunc("POST /logout", h.logout)
	m.HandleFunc("GET /directory", h.auth(h.directory))
	m.HandleFunc("GET /profile", h.auth(h.profile))
	m.HandleFunc("GET /profile/edit", h.auth(h.edit))
	m.HandleFunc("POST /profile/edit", h.auth(h.update))
	m.HandleFunc("GET /switch-company", h.auth(h.switchCompany))
	m.HandleFunc("GET /healthz", h.health)
	st, e := fs.Sub(assets, "web/static")
	if e != nil {
		return nil, e
	}
	m.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(st))))
	return h.log(m), nil
}
func (h *Handler) home(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/directory", 302)
}
func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "version": h.cfg.App.Version})
}
func (h *Handler) loginPage(w http.ResponseWriter, r *http.Request) {
	h.render(w, 200, "login.html", Page{Title: "Вход"})
}
func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	u, e := h.store.Authenticate(r.Context(), r.FormValue("email"), r.FormValue("password"))
	if e != nil {
		h.render(w, 401, "login.html", Page{Title: "Вход", Error: "Неверный email или пароль."})
		return
	}
	token, e := sqlstore.NewToken()
	if e != nil {
		http.Error(w, "session error", 500)
		return
	}
	if e = h.store.CreateSession(r.Context(), u.ID, u.ActiveCompanyID, token); e != nil {
		http.Error(w, "session error", 500)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "cw_session", Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: 28800, Secure: h.cfg.Security.CookieSecure})
	http.Redirect(w, r, "/directory", 302)
}
func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if c, e := r.Cookie("cw_session"); e == nil {
		_ = h.store.DeleteSession(r.Context(), c.Value)
	}
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
		u, e := h.store.Session(r.Context(), c.Value)
		if e != nil {
			http.Redirect(w, r, "/login", 302)
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), userKey{}, u)))
	}
}
func (h *Handler) directory(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(userKey{}).(domain.User)
	es, ds, total, cities, e := h.dir.List(r.Context(), u.ActiveCompanyID, r.URL.Query().Get("q"))
	if e != nil {
		http.Error(w, "directory error", 500)
		return
	}
	h.render(w, 200, "directory.html", Page{Title: "Справочник сотрудников", User: u, Employees: es, Departments: ds, Total: total, Cities: cities})
}
func (h *Handler) profile(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(userKey{}).(domain.User)
	e, err := h.dir.Employee(r.Context(), u.ActiveCompanyID, u.ID)
	if err != nil {
		http.Error(w, "profile not found", 404)
		return
	}
	h.render(w, 200, "profile.html", Page{Title: "Мой профиль", User: u, Employee: &e})
}
func (h *Handler) edit(w http.ResponseWriter, r *http.Request) {
	h.profileTemplate(w, r, "profile_edit.html", "Изменить контакты", "")
}
func (h *Handler) profileTemplate(w http.ResponseWriter, r *http.Request, name, title, msg string) {
	u := r.Context().Value(userKey{}).(domain.User)
	e, err := h.dir.Employee(r.Context(), u.ActiveCompanyID, u.ID)
	if err != nil {
		http.Error(w, "profile not found", 404)
		return
	}
	h.render(w, 200, name, Page{Title: title, User: u, Employee: &e, Error: msg})
}
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(userKey{}).(domain.User)
	v, _ := strconv.Atoi(r.FormValue("version"))
	_, e := h.dir.Update(r.Context(), u.ID, u.ID, repository.ContactUpdate{MobilePhone: r.FormValue("mobile_phone"), PersonalEmail: r.FormValue("personal_email"), ShowPersonalEmail: r.FormValue("show_personal_email") != "", TelegramURL: r.FormValue("telegram_url"), MattermostURL: r.FormValue("mattermost_url"), Version: v})
	if e != nil {
		h.profileTemplate(w, r, "profile_edit.html", "Изменить контакты", "Данные не сохранены: запись изменилась в другой вкладке или доступ запрещён.")
		return
	}
	http.Redirect(w, r, "/profile", 302)
}
func (h *Handler) switchCompany(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(userKey{}).(domain.User)
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	if h.store.SetActiveCompany(r.Context(), u.ID, id) != nil {
		http.Error(w, "forbidden", 403)
		return
	}
	c, _ := r.Cookie("cw_session")
	e := h.store.UpdateSessionCompany(r.Context(), c.Value, id)
	if e != nil {
		http.Error(w, "session error", 500)
		return
	}
	http.Redirect(w, r, "/directory", 302)
}
func (h *Handler) render(w http.ResponseWriter, status int, name string, p Page) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if e := h.t.ExecuteTemplate(w, name, p); e != nil {
		h.logger.Printf("template: %v", e)
	}
}
func (h *Handler) log(n http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := time.Now()
		n.ServeHTTP(w, r)
		h.logger.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(s).Round(time.Millisecond))
	})
}
