package render

import (
	"bytes"
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"sync"
)

// Renderer holds parsed templates keyed by page filename.
type Renderer struct {
	byFile map[string]*template.Template
}

// New parses templates without extra functions (backwards‑compatible).
func New(tfs embed.FS) *Renderer {
	return NewWithFuncs(tfs, template.FuncMap{})
}

// NewWithFuncs parses templates and registers provided helpers.
// British English: allows helpers like asset() for cache busting.
func NewWithFuncs(tfs embed.FS, funcs template.FuncMap) *Renderer {
	// Discover layouts and pages (e.g. base.tmpl.html, navbar.tmpl.html)
	layouts, _ := fs.Glob(tfs, "templates/layouts/*.tmpl.html")
	pages, _ := fs.Glob(tfs, "templates/pages/*.tmpl.html")

	byFile := make(map[string]*template.Template)

	for _, page := range pages {
		name := filepath.Base(page)

		// Start a fresh template per page to avoid cross‑contamination.
		t := template.New(name).Funcs(funcs)

		// Parse layouts first so pages can call {{template "base" .}}.
		t = template.Must(t.ParseFS(tfs, layouts...))

		// Then parse the page itself.
		t = template.Must(t.ParseFS(tfs, page))

		byFile[name] = t
	}

	return &Renderer{byFile: byFile}
}

// HTML renders a named template to the HTTP response safely.
// British English: render into a buffer first, then write headers atomically.
func (r *Renderer) HTML(w http.ResponseWriter, req *http.Request, name string, data any) {
	t, ok := r.byFile[name]
	if !ok {
		http.NotFound(w, req)
		return
	}

	// Use sync.Pool for buffer reuse
	buf := getBuffer()
	defer putBuffer(buf)

	if err := t.ExecuteTemplate(buf, "base", data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}

// Buffer pool for template rendering
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func getBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

func putBuffer(buf *bytes.Buffer) {
	buf.Reset()
	bufferPool.Put(buf)
}
