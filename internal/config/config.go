package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	App       AppConfig       `json:"app"`
	Server    ServerConfig    `json:"server"`
	Database  DatabaseConfig  `json:"database"`
	Paths     PathsConfig     `json:"paths"`
	Security  SecurityConfig  `json:"security"`
	Bootstrap BootstrapConfig `json:"bootstrap"`
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
	Driver   string         `json:"driver"`
	SQLite   SQLiteConfig   `json:"sqlite"`
	Postgres PostgresConfig `json:"postgres"`
	MySQL    MySQLConfig    `json:"mysql"`
}
type SQLiteConfig struct {
	Path string `json:"path"`
}
type PostgresConfig struct {
	DSN string `json:"dsn"`
}
type MySQLConfig struct {
	DSN string `json:"dsn"`
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
type BootstrapConfig struct {
	SeedDemoData bool `json:"seed_demo_data"`
}

func Defaults() Config {
	return Config{App: AppConfig{"Corporate Workspace", "0.2.0-dev"}, Server: ServerConfig{":8080", "http://localhost:8080"}, Database: DatabaseConfig{Driver: "sqlite", SQLite: SQLiteConfig{"data/corporate-workspace.db"}}, Paths: PathsConfig{"data", "logs", "backups", "uploads"}, Bootstrap: BootstrapConfig{true}}
}
func LoadOrCreate(path string) (Config, bool, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return Config{}, false, err
	}
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		cfg := Defaults()
		data, e := json.MarshalIndent(cfg, "", "  ")
		if e != nil {
			return Config{}, false, e
		}
		if e = os.WriteFile(path, data, 0o600); e != nil {
			return Config{}, false, e
		}
		return cfg, true, nil
	}
	if err != nil {
		return Config{}, false, err
	}
	var cfg Config
	if err = json.Unmarshal(raw, &cfg); err != nil {
		return Config{}, false, fmt.Errorf("parse config: %w", err)
	}
	applyDefaults(&cfg)
	return cfg, false, nil
}
func applyDefaults(c *Config) {
	d := Defaults()
	if c.App.Name == "" {
		c.App.Name = d.App.Name
	}
	if c.App.Version == "" {
		c.App.Version = d.App.Version
	}
	if c.Server.Address == "" {
		c.Server.Address = d.Server.Address
	}
	if c.Server.BaseURL == "" {
		c.Server.BaseURL = d.Server.BaseURL
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
}
func EnsureRuntimeDirectories(c Config) error {
	for _, dir := range []string{c.Paths.Data, c.Paths.Logs, c.Paths.Backups, c.Paths.Uploads} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
	}
	return nil
}
