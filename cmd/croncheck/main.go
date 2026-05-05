package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/yourorg/croncheck/internal/config"
)

func main() {
	configPath := flag.String("config", "configs/croncheck.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "croncheck: failed to load config: %v\n", err)
		os.Exit(1)
	}

	logLevel := slog.LevelInfo
	switch cfg.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	slog.Info("croncheck starting",
		"listen_addr", cfg.ListenAddr,
		"job_count", len(cfg.Jobs),
	)

	for _, job := range cfg.Jobs {
		slog.Info("registered job",
			"name", job.Name,
			"schedule", job.Schedule,
			"grace", job.Grace,
		)
	}

	// TODO: start scheduler and HTTP server
	slog.Info("startup complete")
}
