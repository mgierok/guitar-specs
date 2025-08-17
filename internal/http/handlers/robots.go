package handlers

import "net/http"

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
