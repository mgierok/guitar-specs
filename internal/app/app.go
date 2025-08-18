package app

import (
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	h "guitar-specs/internal/http/handlers"
	mw "guitar-specs/internal/http/middleware"
	"guitar-specs/internal/models"
	"guitar-specs/internal/render"
	"guitar-specs/web"
)

// App represents the core application structure and holds all dependencies.
// It encapsulates the HTTP router, configuration, logging, and middleware stack.
type App struct {
	Config Config        // Application configuration (host, port, environment)
	Logger *slog.Logger  // Structured logger for application events
	Router http.Handler  // HTTP router with all middleware and routes configured
	DB     *pgxpool.Pool // PostgreSQL connection pool
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

	// Database connection is mandatory
	dsn := buildPostgresDSN(cfg)
	if dsn == "" {
		logger.Error("database configuration missing; set DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, DB_SSLMODE")
		panic("database configuration missing")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		logger.Error("database pool init failed", "err", err)
		panic(fmt.Sprintf("database init failed: %v", err))
	}
	if err := pool.Ping(ctx); err != nil {
		logger.Error("database ping failed", "err", err)
		panic(fmt.Sprintf("database ping failed: %v", err))
	}
	logger.Info("database connected")

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
	ren := render.New(web.TemplatesFS, template.FuncMap{"asset": assetFunc}, cfg.Env)

	// Create model store and page handlers
	store := models.NewStore(pool)
	pages := h.New(ren, web.RobotsFS, store)

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
		DB:     pool,
	}
}

// buildPostgresDSN assembles a pgx DSN from split parameters if provided; otherwise returns empty string.
func buildPostgresDSN(cfg Config) string {
	if cfg.DBHost == "" || cfg.DBUser == "" || cfg.DBName == "" {
		return ""
	}
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.DBUser, cfg.DBPassword),
		Host:   fmt.Sprintf("%s:%s", cfg.DBHost, cfg.DBPort),
		Path:   "/" + cfg.DBName,
	}
	q := url.Values{}
	if cfg.DBSSLMode != "" {
		q.Set("sslmode", cfg.DBSSLMode)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// Close releases application resources.
func (a *App) Close() {
	if a.DB != nil {
		a.DB.Close()
	}
}
