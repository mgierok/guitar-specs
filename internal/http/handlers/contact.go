package handlers

import "net/http"

func (p *Pages) Contact(w http.ResponseWriter, r *http.Request) {
	p.render.HTML(w, r, "contact.tmpl.html", map[string]any{
		"title": "Contact",
	})
}
