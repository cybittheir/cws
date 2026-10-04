package app

import (
	"context"
	"corporate-workspace/internal/config"
	"corporate-workspace/internal/repository/memory"
	"corporate-workspace/internal/service"
	web "corporate-workspace/internal/transport/http"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type App struct{ server *http.Server }

func New(cfg config.Config, logger *log.Logger) (*App, error) {
	store := memory.New()
	handler, err := web.NewHandler(cfg, logger, store, service.NewDirectory(store))
	if err != nil {
		return nil, err
	}
	return &App{&http.Server{Addr: cfg.Server.Address, Handler: handler}}, nil
}
func (a *App) ListenAndServe() error              { return a.server.ListenAndServe() }
func (a *App) Shutdown(ctx context.Context) error { return a.server.Shutdown(ctx) }
func NewLogger(cfg config.Config) (*log.Logger, func() error, error) {
	if err := os.MkdirAll(cfg.Paths.Logs, 0o755); err != nil {
		return nil, nil, err
	}
	f, err := os.OpenFile(filepath.Join(cfg.Paths.Logs, "application.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, nil, err
	}
	return log.New(io.MultiWriter(os.Stdout, f), "", log.LstdFlags|log.LUTC), f.Close, nil
}
