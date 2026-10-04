package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	App      AppConfig      `json:"app"`
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Paths    PathsConfig    `json:"paths"`
	Security SecurityConfig `json:"security"`
}
type AppConfig struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
type ServerConfig struct {
	Address string `json:"address"`
	BaseURL string `json:"base_url"`
}
type DatabaseConfig struct {
	Driver string `json:"driver"`
	SQLite struct {
		Path string `json:"path"`
	} `json:"sqlite"`
	Postgres struct {
		DSN string `json:"dsn"`
	} `json:"postgres"`
	MySQL struct {
		DSN string `json:"dsn"`
	} `json:"mysql"`
}
type PathsConfig struct {
	Data    string `json:"data"`
	Logs    string `json:"logs"`
	Backups string `json:"backups"`
	Uploads string `json:"uploads"`
}
type SecurityConfig struct {
	CookieSecure bool `json:"cookie_secure"`
}

func Defaults() Config {
	var c Config
	c.App = AppConfig{"Corporate Workspace", "0.2.1-dev"}
	c.Server = ServerConfig{":8080", "http://localhost:8080"}
	c.Database.Driver = "sqlite"
	c.Database.SQLite.Path = "data/corporate-workspace.db"
	c.Paths = PathsConfig{"data", "logs", "backups", "uploads"}
	return c
}
func LoadOrCreate(p string) (Config, bool, error) {
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return Config{}, false, err
	}
	b, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		c := Defaults()
		b, _ = json.MarshalIndent(c, "", "  ")
		if err = os.WriteFile(p, b, 0600); err != nil {
			return Config{}, false, err
		}
		return c, true, nil
	}
	if err != nil {
		return Config{}, false, err
	}
	var c Config
	if err = json.Unmarshal(b, &c); err != nil {
		return c, false, fmt.Errorf("parse config: %w", err)
	}
	d := Defaults()
	if c.Server.Address == "" {
		c.Server.Address = d.Server.Address
	}
	if c.Database.Driver == "" {
		c.Database.Driver = d.Database.Driver
	}
	if c.Database.SQLite.Path == "" {
		c.Database.SQLite.Path = d.Database.SQLite.Path
	}
	if c.Paths.Data == "" {
		c.Paths.Data = d.Paths.Data
	}
	if c.Paths.Logs == "" {
		c.Paths.Logs = d.Paths.Logs
	}
	if c.Paths.Backups == "" {
		c.Paths.Backups = d.Paths.Backups
	}
	if c.Paths.Uploads == "" {
		c.Paths.Uploads = d.Paths.Uploads
	}
	return c, false, nil
}
func EnsureRuntimeDirectories(c Config) error {
	for _, x := range []string{c.Paths.Data, c.Paths.Logs, c.Paths.Backups, c.Paths.Uploads} {
		if err := os.MkdirAll(x, 0755); err != nil {
			return err
		}
	}
	return nil
}
