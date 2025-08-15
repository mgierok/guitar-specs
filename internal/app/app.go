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
	r.Use(middleware.Compress(5))
	r.Use(middleware.Timeout(mw.DefaultTimeout))

	// Static files from embed under /static
	sub, _ := fs.Sub(web.StaticFS, "static")
	fileServer := http.FileServer(http.FS(sub))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// Renderer
	ren := render.New(web.TemplatesFS)

	// Handlers
	pages := h.New(ren, web.RobotsFS)
	r.Get("/", pages.Home)
	r.Get("/about", pages.About)
	r.Get("/contact", pages.Contact)
	r.Get("/robots.txt", pages.RobotsTxt)

	return &App{
		Config: cfg,
		Logger: logger,
		Router: r,
	}
}
