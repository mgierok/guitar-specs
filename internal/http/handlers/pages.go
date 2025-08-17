package handlers

import (
	"embed"
	"net/http"

	"guitar-specs/internal/models"
	"guitar-specs/internal/render"
)

type Pages struct {
	render *render.Renderer
	robots embed.FS
	store  *models.Store
}

func New(r *render.Renderer, robotsFS embed.FS, store *models.Store) *Pages {
	return &Pages{render: r, robots: robotsFS, store: store}
}

func (p *Pages) Home(w http.ResponseWriter, r *http.Request) {
	p.render.HTML(w, r, "home.tmpl.html", map[string]any{
		"title": "Home",
	})
}

func (p *Pages) About(w http.ResponseWriter, r *http.Request) {
	p.render.HTML(w, r, "about.tmpl.html", map[string]any{
		"title": "About Us",
	})
}

func (p *Pages) Contact(w http.ResponseWriter, r *http.Request) {
	p.render.HTML(w, r, "contact.tmpl.html", map[string]any{
		"title": "Contact",
	})
}

func (p *Pages) RobotsTxt(w http.ResponseWriter, r *http.Request) {
	b, err := p.robots.ReadFile("robots.txt")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}

// Guitars renders a simple list of guitars from the database.
func (p *Pages) Guitars(w http.ResponseWriter, r *http.Request) {
	list, err := p.store.Guitars.List(r.Context())
	if err != nil {
		http.Error(w, "Failed to query guitars", http.StatusInternalServerError)
		return
	}
	p.render.HTML(w, r, "guitars.tmpl.html", map[string]any{
		"title":   "Guitars",
		"guitars": list,
	})
}
