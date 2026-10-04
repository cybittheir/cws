package app

import (
	"context"
	"corporate-workspace/internal/config"
	"corporate-workspace/internal/repository/sqlite"
	"corporate-workspace/internal/service"
	web "corporate-workspace/internal/transport/http"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type App struct {
	server *http.Server
	store  *sqlite.Store
}

func New(c config.Config, l *log.Logger) (*App, error) {
	if c.Database.Driver != "sqlite" {
		return nil, fmt.Errorf("stage 2.1 supports sqlite only; configured: %s", c.Database.Driver)
	}
	s, err := sqlite.Open(c.Database.SQLite.Path)
	if err != nil {
		return nil, err
	}
	h, err := web.NewHandler(c, l, s, service.NewDirectory(s))
	if err != nil {
		s.Close()
		return nil, err
	}
	return &App{&http.Server{Addr: c.Server.Address, Handler: h}, s}, nil
}
func (a *App) ListenAndServe() error            { return a.server.ListenAndServe() }
func (a *App) Shutdown(c context.Context) error { return a.server.Shutdown(c) }
func (a *App) Close() error                     { return a.store.Close() }
func NewLogger(c config.Config) (*log.Logger, func() error, error) {
	if err := os.MkdirAll(c.Paths.Logs, 0755); err != nil {
		return nil, nil, err
	}
	f, err := os.OpenFile(filepath.Join(c.Paths.Logs, "application.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, err
	}
	return log.New(io.MultiWriter(os.Stdout, f), "", log.LstdFlags|log.LUTC), f.Close, nil
}
