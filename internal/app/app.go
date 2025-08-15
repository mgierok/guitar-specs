package app

import (
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	h "guitar-specs/internal/http/handlers"
	mw "guitar-specs/internal/http/middleware"
	"guitar-specs/internal/render"
	"guitar-specs/web"
)

// App holds core application state and dependencies.
type App struct {
	Config Config
	Logger *slog.Logger
	Router http.Handler
}

// New wires up the router, middleware, handlers and asset versioning.
func New(cfg Config) *App {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	r := chi.NewRouter()

	// Standard middlewares for all dynamic routes
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(mw.SlogLogger(logger))
	r.Use(chimw.Timeout(mw.DefaultTimeout))

	// Compute per-file hashes for static assets
	sub, _ := fs.Sub(web.StaticFS, "static")
	versions, err := BuildAssetVersions(sub)
	if err != nil {
		logger.Warn("failed to build asset versions", "err", err)
		versions = map[string]string{}
	}

	// Helper function for templates to append ?v=<hash>
	assetFunc := func(p string) string {
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		if v, ok := versions[p]; ok {
			return p + "?v=" + v
		}
		return p
	}

	// Renderer with asset() helper
	ren := render.NewWithFuncs(web.TemplatesFS, template.FuncMap{"asset": assetFunc})

	// Group for static files without verbose request logging
	r.Group(func(r chi.Router) {
		// Long-lived, immutable cache is safe because URLs change when content changes
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				next.ServeHTTP(w, req)
			})
		})

		r.Handle("/static/*", http.StripPrefix("/static/", mw.PrecompressedFileServer(sub)))
	})

	// Pages (dynamic): apply compression only here.
	pages := h.New(ren, web.RobotsFS)
	r.Group(func(r chi.Router) {
		// British English: compress dynamic responses; static assets are handled elsewhere.
		r.Use(chimw.Compress(5,
			"text/html", "text/css", "application/javascript",
			"application/json", "image/svg+xml",
		))

		r.Get("/", pages.Home)
		r.Get("/about", pages.About)
		r.Get("/contact", pages.Contact)
	})

	// Non-compressed, tiny responses
	r.Get("/robots.txt", pages.RobotsTxt)
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
