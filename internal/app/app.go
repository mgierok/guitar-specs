package app

import (
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	h "guitar-specs/internal/http/handlers"
	mw "guitar-specs/internal/http/middleware"
	"guitar-specs/internal/render"
	"guitar-specs/web"
)

// App represents the core application structure and holds all dependencies.
// It encapsulates the HTTP router, configuration, logging, and middleware stack.
type App struct {
	Config Config       // Application configuration (host, port, environment)
	Logger *slog.Logger // Structured logger for application events
	Router http.Handler // HTTP router with all middleware and routes configured
}

// New creates and configures a new application instance.
// It sets up the router, middleware stack, handlers, and asset versioning system.
// The function follows a clear middleware ordering: security → rate limiting → logging → compression.
func New(cfg Config) *App {
	// Create structured logger with text output for development and production
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	r := chi.NewRouter()

	// Standard middleware stack applied to all dynamic routes
	// These provide request identification, security, logging, and timeout protection
	r.Use(chimw.RequestID)                  // Unique request identifier for tracing
	r.Use(chimw.RealIP)                     // Extract real client IP from proxy headers
	r.Use(chimw.Recoverer)                  // Panic recovery and graceful error handling
	r.Use(mw.SlogLogger(logger))            // Structured request logging
	r.Use(chimw.Timeout(mw.DefaultTimeout)) // Request timeout protection
	r.Use(mw.SecurityHeaders)               // Security headers (CSP, XSS protection, etc.)

	// Rate limiting: 100 requests per minute per IP address
	// This protects against abuse and ensures fair resource distribution
	rateLimiter := mw.NewRateLimiter(100, time.Minute)
	r.Use(rateLimiter.RateLimit)

	// Compute per-file hashes for static assets to enable cache busting
	// This ensures clients always receive the latest version when assets change
	sub, _ := fs.Sub(web.StaticFS, "static")
	versions, err := BuildAssetVersions(sub)
	if err != nil {
		logger.Warn("failed to build asset versions", "err", err)
		versions = map[string]string{}
	}

	// Helper function for templates to append version hash to asset URLs
	// This enables aggressive caching while ensuring cache invalidation on updates
	assetFunc := func(p string) string {
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		if v, ok := versions[p]; ok {
			return p + "?v=" + v
		}
		return p
	}

	// Create renderer with asset versioning helper function
	// Templates can now use {{ asset "/static/css/main.css" }} for cache-busted URLs
	ren := render.NewWithFuncs(web.TemplatesFS, template.FuncMap{"asset": assetFunc})

	// Static file serving group with aggressive caching
	// These files are served without verbose logging and with long-lived cache headers
	r.Group(func(r chi.Router) {
		// Long-lived, immutable cache is safe because URLs change when content changes
		// This enables maximum browser caching for static assets
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				next.ServeHTTP(w, req)
			})
		})

		// Serve static files with intelligent compression and caching
		// PrecompressedFileServer automatically selects the best compression format
		r.Handle("/static/*", http.StripPrefix("/static/", mw.PrecompressedFileServer(sub)))
	})

	// Dynamic page routes with compression and caching optimisations
	// These routes generate HTML content that benefits from compression and ETag caching
	pages := h.New(ren, web.RobotsFS)
	r.Group(func(r chi.Router) {
		// Add ETag for better caching (BEFORE compression)
		// ETag middleware must run before compression to ensure headers are set correctly
		r.Use(mw.ETag)

		// Apply compression to dynamic responses for bandwidth reduction
		// Static assets are handled separately with precompression
		r.Use(chimw.Compress(5,
			"text/html", "text/css", "application/javascript",
			"application/json", "image/svg+xml",
		))

		// Page routes that generate dynamic HTML content
		r.Get("/", pages.Home)           // Homepage
		r.Get("/about", pages.About)     // About page
		r.Get("/contact", pages.Contact) // Contact page
	})

	// Utility endpoints that don't require compression
	// These are small, frequently accessed responses
	r.Get("/robots.txt", pages.RobotsTxt) // Search engine crawling instructions
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	return &App{
		Config: cfg,
		Logger: logger,
		Router: r,
	}
}
