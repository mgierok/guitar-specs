package main

import (
	"context"
	"errors"
	"log"
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

func main() {
	// Load configuration from environment variables and .env files
	cfg := app.LoadConfig()

	// Validate HTTPS configuration
	if err := cfg.ValidateHTTPS(); err != nil {
		log.Fatalf("HTTPS configuration error: %v", err)
	}

	a := app.New(cfg)
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
		log.Printf("HTTPS server starting on %s", cfg.Addr())
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
			log.Fatalf("HTTPS server error: %v", err)
		}
	case <-quit:
		// proceed to graceful shutdown below
	}

	log.Println("shutting down HTTPS server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Graceful shutdown with timeout
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	} else {
		log.Println("server shutdown completed successfully")
	}

	// Force close if shutdown timeout reached
	select {
	case <-shutdownCtx.Done():
		log.Println("shutdown timeout reached, forcing exit")
		if err := srv.Close(); err != nil {
			log.Printf("force close error: %v", err)
		}
	default:
		log.Println("all servers stopped gracefully")
	}
}
