package handlers

import "net/http"

func (p *Pages) Home(w http.ResponseWriter, r *http.Request) {
	// Set content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Render template using new interface with request context
	if err := p.render.RenderWithRequest(w, "home", r, map[string]any{
		"Title": "Home",
	}); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
