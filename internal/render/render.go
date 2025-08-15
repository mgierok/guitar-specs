package render

import (
	"bytes"
	"embed"
	"html/template"
	"net/http"
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
	t, ok := r.byFile[name]
	if !ok {
		http.NotFound(w, req)
		return
	}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "base", data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}
