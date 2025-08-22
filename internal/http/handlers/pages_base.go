package handlers

import (
	"embed"

	"guitar-specs/internal/models"
	"guitar-specs/internal/render"
)

// Pages groups page handlers and shared dependencies.
type Pages struct {
	render render.Renderer
	robots embed.FS
	store  *models.Store
}

// New constructs a Pages handler set.
func New(r render.Renderer, robotsFS embed.FS, store *models.Store) *Pages {
	return &Pages{render: r, robots: robotsFS, store: store}
}
