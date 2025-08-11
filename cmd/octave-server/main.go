package main

import (
	"flag"
	"log"
	"log/slog"
	"os"

	"github.com/fmcato/octave-mcp/internal/server"
)

var httpAddr = flag.String("http", "", "HTTP address to listen on (empty for stdio)")

func main() {
	// Setup structured logging
	logLevel := slog.LevelInfo
	if envLevel := os.Getenv("LOG_LEVEL"); envLevel != "" {
		switch envLevel {
		case "debug":
			logLevel = slog.LevelDebug
		case "info":
			logLevel = slog.LevelInfo
		case "warn":
			logLevel = slog.LevelWarn
		case "error":
			logLevel = slog.LevelError
		}
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting octave-server")
	defer slog.Info("Shutting down octave-server")

	flag.Parse()

	srv := server.New()
	srv.RegisterHandlers()

	if *httpAddr != "" {
		if err := srv.RunHTTP(*httpAddr); err != nil {
			slog.Error("HTTP server failed", "error", err)
			log.Fatal(err)
		}
	}
	if err := srv.RunStdio(); err != nil {
		slog.Error("Stdio server failed", "error", err)
		log.Fatal(err)
	}

}
