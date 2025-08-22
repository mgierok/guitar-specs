package handlers

import "net/http"

func (p *Pages) Contact(w http.ResponseWriter, r *http.Request) {
	// Set content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Render template using new interface with request context
	if err := p.render.RenderWithRequest(w, "contact", r, map[string]any{
		"Title": "Contact",
	}); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
