package handlers

import "net/http"

// Guitars renders a simple list of guitars from the database.
func (p *Pages) Guitars(w http.ResponseWriter, r *http.Request) {
	list, err := p.store.Guitars.List(r.Context())
	if err != nil {
		http.Error(w, "Failed to query guitars", http.StatusInternalServerError)
		return
	}
	// Set content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Render template using new interface with request context
	if err := p.render.RenderWithRequest(w, "guitars", r, map[string]any{
		"Title":   "Guitars",
		"guitars": list,
	}); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
