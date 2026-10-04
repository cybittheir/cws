package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"corporate-workspace/internal/app"
	"corporate-workspace/internal/config"
)

func main() {
	configPath := flag.String("config", "config/config.json", "path to JSON configuration")
	initOnly := flag.Bool("init", false, "create runtime directories and configuration, then exit")
	flag.Parse()

	cfg, created, err := config.LoadOrCreate(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	if created {
		log.Printf("created configuration file: %s", *configPath)
	}
	if err := config.EnsureRuntimeDirectories(cfg); err != nil {
		log.Fatal(err)
	}
	if *initOnly {
		log.Println("initialization completed")
		return
	}

	logger, closeLog, err := app.NewLogger(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer closeLog()

	application, err := app.New(cfg, logger)
	if err != nil {
		logger.Fatal(err)
	}
	go func() {
		logger.Printf("starting %s on %s", cfg.App.Name, cfg.Server.Address)
		logger.Printf("open http://localhost%s", cfg.Server.Address)
		if err := application.ListenAndServe(); err != nil && err.Error() != "http: Server closed" {
			logger.Fatal(err)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := application.Shutdown(ctx); err != nil {
		logger.Printf("shutdown error: %v", err)
	}
}
