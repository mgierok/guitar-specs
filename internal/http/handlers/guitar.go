package handlers

import (
	"net/http"
	"strings"
)

// GuitarDetail renders a single guitar with its features.
// Path expected: /guitar/{slug}
func (p *Pages) GuitarDetail(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/guitar/")
	slug := strings.TrimSuffix(path, "/")
	if slug == "" || strings.Contains(slug, "/") {
		http.NotFound(w, r)
		return
	}

	g, err := p.store.Guitars.GetBySlug(r.Context(), slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	feats, err := p.store.Guitars.ListFeaturesBySlug(r.Context(), slug)
	if err != nil {
		http.Error(w, "Failed to load features", http.StatusInternalServerError)
		return
	}

	// Attach features to the guitar
	g.Features = feats

	// Set content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Render template using new interface with request context
	if err := p.render.RenderWithRequest(w, "guitar", r, map[string]any{
		"Title":  g.BrandName + " " + g.Model,
		"guitar": g,
	}); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
