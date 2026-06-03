package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"thermalik-bridge/internal/bridge"
	"thermalik-bridge/internal/config"
	"thermalik-bridge/internal/httpapi"
	"thermalik-bridge/internal/store"
	"thermalik-bridge/internal/transport"
	"thermalik-bridge/internal/updater"
)

func cmdRun(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	configPath := fs.String("config", "", "Path to the thermalik-bridge JSON config file.")
	listen := fs.String("listen", "", "Override the listen address.")
	_ = fs.Parse(args)

	cfgPath, err := resolveConfigPath(*configPath)
	if err != nil {
		log.Fatalf("resolve config path: %v", err)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if *listen != "" {
		cfg.Server.Address = *listen
	}

	if err := os.MkdirAll(filepath.Dir(cfg.Storage.JobStorePath), 0o755); err != nil {
		log.Fatalf("create storage directory: %v", err)
	}

	jobStore, err := store.NewFileJobStore(cfg.Storage.JobStorePath)
	if err != nil {
		log.Fatalf("create job store: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Non-blocking update check on startup.
	go func() {
		info, err := updater.CheckLatest()
		if err != nil || !info.Available {
			return
		}
		log.Printf("update available: %s → run 'thermalik-bridge update' to upgrade", info.Version)
	}()

	service, err := bridge.NewService(ctx, cfg, jobStore, transport.NewDefaultManager())
	if err != nil {
		log.Fatalf("create bridge service: %v", err)
	}

	server := &http.Server{
		Addr:              cfg.Server.Address,
		Handler:           httpapi.NewHandler(service, cfg.Server.AllowedOrigins),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("shutdown error: %v", err)
		}
	}()

	log.Printf("thermalik-bridge listening on http://%s using %s", cfg.Server.Address, cfgPath)

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("listen: %v", err)
	}
}

func resolveConfigPath(flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}

	return config.DefaultConfigPath()
}
