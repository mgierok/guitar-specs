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
	"guitar-specs/internal/config"
	"guitar-specs/internal/db"
	"guitar-specs/internal/assets"
	"guitar-specs/internal/render"
	"guitar-specs/web"
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
	// Create startup logger with full logging (always INFO level)
	startupLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	startupLogger.Info("application starting")

	// 1. Load configuration using new config package
	configProvider := config.New()
	cfg := configProvider.Get()
	
	if err := configProvider.Validate(); err != nil {
		startupLogger.Error("configuration validation failed", "error", err)
		os.Exit(1)
	}

	// Create runtime logger with configurable level from environment
	runtimeLogger := setupLogger(cfg.LogLevel)

	// 2. Validate HTTPS configuration
	if err := cfg.ValidateHTTPS(); err != nil {
		startupLogger.Error("HTTPS configuration error", "error", err)
		os.Exit(1)
	}

	startupLogger.Info("configuration loaded successfully", "log_level", cfg.LogLevel, "env", cfg.Env)

	// 3. Initialize database connection
	startupLogger.Info("initializing database connection")
	dbConfig := db.DatabaseConfig{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		Database: cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	}
	
	database := db.New(dbConfig)
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := database.Connect(ctx); err != nil {
		startupLogger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	
	if err := database.Ping(ctx); err != nil {
		startupLogger.Error("database ping failed", "error", err)
		os.Exit(1)
	}
	
	startupLogger.Info("database connected successfully")
	defer database.Close()

	// 4. Initialize asset manager
	startupLogger.Info("initializing asset manager")
	assetManager, err := assets.New(web.StaticFS, runtimeLogger)
	if err != nil {
		startupLogger.Error("asset manager initialization failed", "error", err)
		os.Exit(1)
	}
	startupLogger.Info("asset manager initialized successfully")

	// 5. Initialize template renderer
	startupLogger.Info("initializing template renderer")
	templateRenderer, err := render.New(web.TemplatesFS, assetManager, cfg.Env, runtimeLogger)
	if err != nil {
		startupLogger.Error("template renderer initialization failed", "error", err)
		os.Exit(1)
	}
	startupLogger.Info("template renderer initialized successfully")

	// 6. Create application with all dependencies
	startupLogger.Info("creating application instance")
	a := app.New(cfg, runtimeLogger, database, templateRenderer)
	defer a.Close()

	startupLogger.Info("application instance created successfully")

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
