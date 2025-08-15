package handlers

import (
	"embed"
	"io/fs"
	"net/http"

	"guitar-specs/internal/render"
)

type Pages struct {
	render  *render.Renderer
	robots  embed.FS
}

func New(r *render.Renderer, robotsFS embed.FS) *Pages {
	return &Pages{render: r, robots: robotsFS}
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
	b, err := fs.ReadFile(p.robots, "robots.txt")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}
