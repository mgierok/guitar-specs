package app

import (
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	h "guitar-specs/internal/http/handlers"
	mw "guitar-specs/internal/http/middleware"
	"guitar-specs/internal/render"
	"guitar-specs/web" // Importing web package for FS variables
)

func New(cfg Config) *App {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	r := chi.NewRouter()

	// Standard middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(mw.SlogLogger(logger))
	r.Use(middleware.RealIP)
	r.Use(middleware.Compress(5))
	r.Use(middleware.Timeout(mw.DefaultTimeout))
	r.Use(middleware.Compress(5,
		"text/html", "text/css", "application/javascript",
		"application/json", "image/svg+xml",
	))

	// Group for static files without verbose request logging
	r.Group(func(r chi.Router) {
		// very long cache for static assets (safe when files are fingerprinted)
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				// Keep this header here if you also serve non-precompressed originals
				w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
				next.ServeHTTP(w, req)
			})
		})

		sub, _ := fs.Sub(web.StaticFS, "static")

		// Important: StripPrefix before the precompressed file server
		r.Handle("/static/*", http.StripPrefix("/static/", mw.PrecompressedFileServer(sub)))
	})

	// Renderer
	ren := render.New(web.TemplatesFS)

	// Handlers
	pages := h.New(ren, web.RobotsFS)
	r.Get("/", pages.Home)
	r.Get("/about", pages.About)
	r.Get("/contact", pages.Contact)
	r.Get("/robots.txt", pages.RobotsTxt)
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	return &App{
		Config: cfg,
		Logger: logger,
		Router: r,
	}
}

func cacheForever(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// jeśli masz fingerprint w nazwach plików, śmiało immutable
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
}
