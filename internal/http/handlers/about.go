package handlers

import "net/http"

func (p *Pages) About(w http.ResponseWriter, r *http.Request) {
	p.render.HTML(w, r, "about.tmpl.html", map[string]any{
		"title": "About Us",
	})
}
