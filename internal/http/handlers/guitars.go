package handlers

import "net/http"

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
