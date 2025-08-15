package render

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

type Renderer struct {
	byFile map[string]*template.Template
}

func New(fs embed.FS) *Renderer {
	base := template.Must(template.New("base").ParseFS(fs, "templates/layouts/*.html"))

	build := func(name string) *template.Template {
		t := template.Must(base.Clone())
		template.Must(t.ParseFS(fs, "templates/pages/"+name))
		return t
	}

	return &Renderer{
		byFile: map[string]*template.Template{
			"home.tmpl.html":    build("home.tmpl.html"),
			"about.tmpl.html":   build("about.tmpl.html"),
			"contact.tmpl.html": build("contact.tmpl.html"),
		},
	}
}

func (r *Renderer) HTML(w http.ResponseWriter, req *http.Request, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t := r.byFile[filepath.Base(name)]
	if t == nil {
		http.NotFound(w, req)
		return
	}
	if err := t.ExecuteTemplate(w, filepath.Base(name), data); err != nil {
		log.Printf("template execute error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
