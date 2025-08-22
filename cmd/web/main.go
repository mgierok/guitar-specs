package main

import (
	"context"
	"errors"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"guitar-specs/internal/app"
)

// ensure consistent MIME types for JavaScript assets across environments.
func init() {
	// application/javascript is the modern, widely expected type for .js
	_ = mime.AddExtensionType(".js", "application/javascript")
	// some bundlers emit .mjs; text/javascript keeps older agents happy
	_ = mime.AddExtensionType(".mjs", "text/javascript")
}

// setupLogger creates a logger with the specified level for runtime operations
func setupLogger(level string) *slog.Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
}

func main() {
	// Load configuration from environment variables and .env files
	cfg := app.LoadConfig()

	// Create startup logger with full logging (always INFO level)
	startupLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create runtime logger with configurable level from environment
	runtimeLogger := setupLogger(cfg.LogLevel)

	// Validate HTTPS configuration
	if err := cfg.ValidateHTTPS(); err != nil {
		startupLogger.Error("HTTPS configuration error", "error", err)
		os.Exit(1)
	}

	startupLogger.Info("application starting", "log_level", cfg.LogLevel, "env", cfg.Env)

	a := app.New(cfg, runtimeLogger)
	defer a.Close()

	// Create HTTPS server
	srv := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           a.Router,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
	}

	// Start HTTPS server
	serverErr := make(chan error, 1)
	go func() {
		startupLogger.Info("HTTPS server starting", "addr", cfg.Addr())
		if err := srv.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile); !errors.Is(err, http.ErrServerClosed) {
			// Propagate non-shutdown errors to the main goroutine so we can fail fast
			serverErr <- err
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Either the server failed (startup/runtime) or we received a shutdown signal
	select {
	case err := <-serverErr:
		if err != nil { // Fail fast on unexpected server errors
			startupLogger.Error("HTTPS server error", "error", err)
			os.Exit(1)
		}
	case <-quit:
		// proceed to graceful shutdown below
	}

	startupLogger.Info("shutting down HTTPS server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Graceful shutdown with timeout
	if err := srv.Shutdown(shutdownCtx); err != nil {
		startupLogger.Error("server shutdown error", "error", err)
	} else {
		startupLogger.Info("server shutdown completed successfully")
	}

	// Force close if shutdown timeout reached
	select {
	case <-shutdownCtx.Done():
		startupLogger.Warn("shutdown timeout reached, forcing exit")
		if err := srv.Close(); err != nil {
			startupLogger.Error("force close error", "error", err)
		}
	default:
		startupLogger.Info("all servers stopped gracefully")
	}
}
