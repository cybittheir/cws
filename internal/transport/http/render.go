package http

import (
	"net/http"
	"time"
)

func (h *Handler) render(w http.ResponseWriter, name string, data Page) {
	h.renderStatus(w, 200, name, data)
}
func (h *Handler) renderStatus(w http.ResponseWriter, status int, name string, data Page) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := h.templates.ExecuteTemplate(w, name, data); err != nil {
		h.logger.Printf("template error: %v", err)
	}
}
func (h *Handler) logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := time.Now()
		next.ServeHTTP(w, r)
		h.logger.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(s).Round(time.Millisecond))
	})
}
