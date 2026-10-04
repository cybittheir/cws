package http

import (
	"context"
	"corporate-workspace/internal/domain/identity"
	"corporate-workspace/internal/repository"
	"net/http"
	"strconv"
)

type userKey struct{}

func contextWithUser(ctx context.Context, u identity.User) context.Context {
	return context.WithValue(ctx, userKey{}, u)
}
func userFrom(r *http.Request) identity.User { return r.Context().Value(userKey{}).(identity.User) }
func (h *Handler) directoryPage(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	companies, _ := h.store.CompaniesForUser(r.Context(), u.ID)
	result, err := h.directory.List(r.Context(), u.ActiveCompanyID, repository.DirectoryFilter{Query: r.URL.Query().Get("q"), Page: 1, PerPage: 50})
	if err != nil {
		http.Error(w, "directory error", 500)
		return
	}
	company := activeCompany(companies, u.ActiveCompanyID)
	h.render(w, "directory.html", Page{Title: "Справочник сотрудников", User: u, Companies: companies, Company: company, Result: result, ShowCity: result.DistinctCities > 1})
}
func (h *Handler) profile(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	e, err := h.directory.Employee(r.Context(), u.ActiveCompanyID, u.ID)
	if err != nil {
		http.Error(w, "profile not found", 404)
		return
	}
	h.render(w, "profile.html", Page{Title: "Мой профиль", User: u, Employee: &e})
}
func (h *Handler) editProfilePage(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	e, err := h.directory.Employee(r.Context(), u.ActiveCompanyID, u.ID)
	if err != nil {
		http.Error(w, "profile not found", 404)
		return
	}
	h.render(w, "profile_edit.html", Page{Title: "Изменить контакты", User: u, Employee: &e})
}
func (h *Handler) updateProfile(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	version, _ := strconv.Atoi(r.FormValue("version"))
	e, err := h.directory.UpdateOwnContacts(r.Context(), u.ID, u.ID, repository.ContactUpdate{MobilePhone: r.FormValue("mobile_phone"), PersonalEmail: r.FormValue("personal_email"), ShowPersonalEmail: r.FormValue("show_personal_email") != "", TelegramURL: r.FormValue("telegram_url"), MattermostURL: r.FormValue("mattermost_url"), Version: version})
	if err != nil {
		h.renderStatus(w, 409, "profile_edit.html", Page{Title: "Изменить контакты", User: u, Employee: &e, Error: "Не удалось сохранить: данные были изменены или доступ запрещён."})
		return
	}
	http.Redirect(w, r, "/profile", 302)
}
func (h *Handler) switchCompany(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	for _, m := range u.Memberships {
		if m.CompanyID == id && m.Active {
			u.ActiveCompanyID = id
			http.SetCookie(w, &http.Cookie{Name: "cw_company", Value: strconv.Itoa(id), Path: "/", MaxAge: 28800, HttpOnly: true})
			break
		}
	}
	http.Redirect(w, r, "/directory", 302)
}
func activeCompany(xs []identity.Membership, id int) identity.Membership {
	for _, x := range xs {
		if x.CompanyID == id {
			return x
		}
	}
	return identity.Membership{}
}
