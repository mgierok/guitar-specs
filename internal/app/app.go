package app

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"guitar-specs/internal/config"
	"guitar-specs/internal/db"
	h "guitar-specs/internal/http/handlers"
	mw "guitar-specs/internal/http/middleware"
	"guitar-specs/internal/models"
	"guitar-specs/internal/render"
	"guitar-specs/web"
)

// App represents the core application structure and holds all dependencies.
// It encapsulates the HTTP router, configuration, logging, and middleware stack.
type App struct {
	Config *config.AppConfig // Application configuration (host, port, environment)
	Logger *slog.Logger      // Structured logger for application events
	Router http.Handler      // HTTP router with all middleware and routes configured
	DB     *pgxpool.Pool     // PostgreSQL connection pool
}

// New creates a new application instance with pre-initialized dependencies.
// This function allows for better dependency injection and testing.
func New(cfg *config.AppConfig, logger *slog.Logger, database db.DatabaseProvider, renderer render.Renderer) *App {
	// Initialize standard Go 1.22 router with pattern matching
	mux := http.NewServeMux()

	// Prepare static file system for serving
	sub, _ := fs.Sub(web.StaticFS, "static")

	// Create model store and page handlers
	store := models.NewStore(database.GetPool())
	pages := h.New(renderer, web.RobotsFS, store)

	// Static file serving with aggressive caching
	// These files are served with long-lived cache headers
	staticHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Long-lived, immutable cache is safe because URLs change when content changes
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		http.FileServer(http.FS(sub)).ServeHTTP(w, r)
	})

	// Create page handlers (no compression)
	homeHandler := http.HandlerFunc(pages.Home)
	aboutHandler := http.HandlerFunc(pages.About)
	contactHandler := http.HandlerFunc(pages.Contact)

	// Register routes with Go 1.22+ pattern matching
	// This provides automatic 405 Method Not Allowed and Allow headers
	// Order matters: more specific patterns first, then general ones
	mux.Handle("/static/", http.StripPrefix("/static/", staticHandler))
	mux.Handle("GET /about", aboutHandler)
	mux.Handle("GET /contact", contactHandler)
	mux.Handle("GET /robots.txt", http.HandlerFunc(pages.RobotsTxt))
	mux.Handle("GET /guitars", http.HandlerFunc(pages.Guitars))
	mux.Handle("GET /guitar/", http.HandlerFunc(pages.GuitarDetail))
	mux.Handle("GET /healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	// Root path without pattern matching to avoid conflicts with /static/
	mux.Handle("/", homeHandler)

	// Apply middleware stack to all routes
	// Order is critical: RequestID → RealIP → Recoverer → Logging → Timeout → Security
	handler := mw.RequestID(
		mw.RealIP(cfg.TrustedProxies)(
			mw.Recoverer(logger)(
				mw.SlogLogger(logger)(
					mw.TimeoutWithCause(mw.DefaultTimeout, fmt.Errorf("request timeout after %v", mw.DefaultTimeout))(
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
		DB:     database.GetPool(),
	}
}

// Close releases application resources.
func (a *App) Close() {
	if a.DB != nil {
		a.DB.Close()
	}
}
