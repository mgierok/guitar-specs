package handlers

import "net/http"

func (p *Pages) Home(w http.ResponseWriter, r *http.Request) {
	p.render.HTML(w, r, "home.tmpl.html", map[string]any{
		"title": "Home",
	})
}
