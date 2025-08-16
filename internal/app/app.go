package app

import (
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

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
// The function follows a clear middleware ordering: security → logging → timeout → compression.
// All middleware is thread-safe and designed for concurrent use.
func New(cfg Config) *App {
	// Create structured logger with text output for development and production
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Initialize standard Go 1.22 router with pattern matching
	mux := http.NewServeMux()

	// Compute per-file hashes for static assets to enable cache busting
	// This ensures clients always receive the latest version when assets change
	// Application requires all assets to be processed successfully for stability
	sub, _ := fs.Sub(web.StaticFS, "static")
	versions, err := BuildAssetVersions(sub)
	if err != nil {
		logger.Warn("failed to build asset versions, retrying once", "err", err)
		// Retry once after a short delay
		time.Sleep(100 * time.Millisecond)
		if versions, err = BuildAssetVersions(sub); err != nil {
			logger.Error("failed to build asset versions after retry, application cannot start", "err", err)
			// Panic to prevent application from running with incomplete assets
			panic(fmt.Sprintf("asset versioning failed: %v", err))
		} else {
			logger.Info("asset versions built successfully after retry")
		}
	} else {
		logger.Info("asset versions built successfully", "count", len(versions))
		// Log information about large files that were skipped
		if len(versions) > 0 {
			logger.Debug("asset versioning completed", "processed_files", len(versions), "max_file_size", "10MB")
		}
	}

	// Helper function for templates to append version hash to asset URLs
	// This enables aggressive caching while ensuring cache invalidation on updates
	// The function is thread-safe as it only reads from the versions map
	assetFunc := func(p string) string {
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		// Read-only access to versions map is safe for concurrent use
		if v, ok := versions[p]; ok {
			return p + "?v=" + v
		}
		return p
	}

	// Create renderer with asset versioning helper function
	// Templates can now use {{ asset "/static/css/main.css" }} for cache-busted URLs
	ren := render.NewWithFuncs(web.TemplatesFS, template.FuncMap{"asset": assetFunc})

	// Create page handlers
	pages := h.New(ren, web.RobotsFS)

	// Static file serving with aggressive caching
	// These files are served with long-lived cache headers
	staticHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Long-lived, immutable cache is safe because URLs change when content changes
		// This enables maximum browser caching for static assets
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		mw.PrecompressedFileServer(sub).ServeHTTP(w, r)
	})

	// Dynamic page routes with compression and caching optimisations
	// These routes generate HTML content that benefits from compression and ETag caching
	dynamicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add ETag for better caching (BEFORE compression)
		// ETag middleware must run before compression to ensure headers are set correctly
		etagHandler := mw.ETag(
			// Apply compression to dynamic responses for bandwidth reduction
			mw.Compress(5,
				"text/html", "text/css", "application/javascript",
				"application/json", "image/svg+xml",
			)(
				// Route to appropriate page handler
				routePageHandler(pages, w, r),
			),
		)
		etagHandler.ServeHTTP(w, r)
	})

	// Register routes with pattern matching
	mux.Handle("/static/", http.StripPrefix("/static/", staticHandler))
	mux.Handle("/", dynamicHandler)
	mux.Handle("/about", dynamicHandler)
	mux.Handle("/contact", dynamicHandler)

	// Utility endpoints that don't require compression
	// These are small, frequently accessed responses
	mux.HandleFunc("/robots.txt", pages.RobotsTxt) // Search engine crawling instructions
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Apply middleware stack to all routes
	// Order is critical: RequestID → RealIP → Recoverer → Logging → Timeout → Security
	handler := mw.RequestID(
		mw.RealIP(cfg.TrustedProxies)(
			mw.Recoverer(logger)(
				mw.SlogLogger(logger)(
					mw.Timeout(mw.DefaultTimeout)(
						mw.SecurityHeaders(mux),
					),
				),
			),
		),
	)

	return &App{
		Config: cfg,
		Logger: logger,
		Router: handler,
	}
}

// routePageHandler routes page requests to the appropriate handler.
func routePageHandler(pages *h.Pages, w http.ResponseWriter, r *http.Request) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			pages.Home(w, r)
		case "/about":
			pages.About(w, r)
		case "/contact":
			pages.Contact(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}
