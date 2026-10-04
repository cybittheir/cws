package main

import (
	"context"
	"corporate-workspace/internal/app"
	"corporate-workspace/internal/config"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	p := flag.String("config", "config/config.json", "path to JSON configuration")
	initOnly := flag.Bool("init", false, "create runtime directories and config, then exit")
	flag.Parse()
	cfg, created, err := config.LoadOrCreate(*p)
	if err != nil {
		log.Fatal(err)
	}
	if created {
		log.Printf("created configuration: %s", *p)
	}
	if err = config.EnsureRuntimeDirectories(cfg); err != nil {
		log.Fatal(err)
	}
	if *initOnly {
		return
	}
	logger, closeLog, err := app.NewLogger(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer closeLog()
	a, err := app.New(cfg, logger)
	if err != nil {
		logger.Fatal(err)
	}
	defer a.Close()
	go func() {
		logger.Printf("starting %s on %s", cfg.App.Name, cfg.Server.Address)
		if err := a.ListenAndServe(); err != nil && err.Error() != "http: Server closed" {
			logger.Fatal(err)
		}
	}()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = a.Shutdown(ctx)
}
