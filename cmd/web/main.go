package main

import (
	"context"
	"errors"
	"log"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"sync"
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

	// Validate HTTPS configuration if enabled
	if err := cfg.ValidateHTTPS(); err != nil {
		log.Fatalf("HTTPS configuration error: %v", err)
	}

	a := app.New(cfg)

	// Create servers slice to track all running servers for graceful shutdown
	var servers []*http.Server

	// Create HTTPS server when enabled
	var srv *http.Server
	if cfg.EnableHTTPS {
		srv = &http.Server{
			Addr:              cfg.AddrHTTPS(),
			Handler:           a.Router,
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
			ReadTimeout:       cfg.ReadTimeout,
			WriteTimeout:      cfg.WriteTimeout,
			IdleTimeout:       cfg.IdleTimeout,
			MaxHeaderBytes:    cfg.MaxHeaderBytes,
		}
		servers = append(servers, srv)

		// Start HTTPS server
		go func() {
			log.Printf("HTTPS server starting on %s", cfg.AddrHTTPS())
			if err := srv.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile); !errors.Is(err, http.ErrServerClosed) {
				log.Printf("HTTPS server error: %v", err)
			}
		}()

		// Start HTTP redirect server if redirect is enabled
		if cfg.RedirectHTTP {
			log.Printf("Configuring HTTP redirect server on %s", cfg.AddrHTTP())

			redirectHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Redirect to HTTPS server on the configured HTTPS port
				httpsURL := "https://" + cfg.Host + ":" + cfg.Port + r.RequestURI
				log.Printf("Redirecting HTTP request from %s to %s", r.URL.String(), httpsURL)
				http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
			})

			redirectSrv := &http.Server{
				Addr:              cfg.AddrHTTP(),
				Handler:           redirectHandler,
				ReadHeaderTimeout: cfg.ReadHeaderTimeout,
				ReadTimeout:       cfg.ReadTimeout,
				WriteTimeout:      cfg.WriteTimeout,
				IdleTimeout:       cfg.IdleTimeout,
			}
			servers = append(servers, redirectSrv)

			go func() {
				log.Printf("HTTP redirect server starting on %s", cfg.AddrHTTP())
				if err := redirectSrv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
					log.Printf("HTTP redirect server error: %v", err)
				}
			}()
		} else {
			log.Printf("HTTP redirect disabled (REDIRECT_HTTP=false)")
		}
	} else {
		// HTTP-only server
		srv = &http.Server{
			Addr:              cfg.Addr(),
			Handler:           a.Router,
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
			ReadTimeout:       cfg.ReadTimeout,
			WriteTimeout:      cfg.WriteTimeout,
			IdleTimeout:       cfg.IdleTimeout,
			MaxHeaderBytes:    cfg.MaxHeaderBytes,
		}
		servers = append(servers, srv)

		// Start HTTP server
		go func() {
			log.Printf("HTTP server starting on %s", cfg.Addr())
			if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
				log.Printf("HTTP server error: %v", err)
			}
		}()
	}

	// SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	// Graceful shutdown of all servers
	log.Println("shutting down servers...")

	// Create shutdown context with proper timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Use WaitGroup to ensure all servers are properly shutdown
	var wg sync.WaitGroup

	// Shutdown all servers concurrently with individual timeouts
	for _, server := range servers {
		wg.Add(1)
		go func(s *http.Server) {
			defer wg.Done()

			// Individual server shutdown with shorter timeout
			serverCtx, serverCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer serverCancel()

			if err := s.Shutdown(serverCtx); err != nil {
				log.Printf("server shutdown error: %v", err)
			} else {
				log.Printf("server shutdown completed successfully")
			}
		}(server)
	}

	// Wait for all servers to shutdown with overall timeout
	shutdownDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(shutdownDone)
	}()

	// Wait for shutdown completion or timeout
	select {
	case <-shutdownDone:
		log.Println("all servers stopped gracefully")
	case <-shutdownCtx.Done():
		log.Println("shutdown timeout reached, forcing exit")
		// Force close any remaining connections
		for _, server := range servers {
			if err := server.Close(); err != nil {
				log.Printf("force close error: %v", err)
			}
		}
	}
}
