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

// Renderer manages HTML template parsing and rendering with performance optimisations.
// It uses a buffer pool to reduce memory allocations and parses templates
// at startup to avoid runtime parsing overhead.
type Renderer struct {
	byFile map[string]*template.Template // Maps page names to parsed templates
}

// New creates a renderer with basic template parsing capabilities.
// This function maintains backwards compatibility for existing code.
func New(tfs embed.FS) *Renderer {
	return NewWithFuncs(tfs, template.FuncMap{})
}

// NewWithFuncs creates a renderer with custom template helper functions.
// This allows templates to use functions like asset() for cache busting URLs.
func NewWithFuncs(tfs embed.FS, funcs template.FuncMap) *Renderer {
	// Discover template files in the embedded filesystem
	// Layouts provide the base structure, pages provide the content
	layouts, _ := fs.Glob(tfs, "templates/layouts/*.tmpl.html")
	pages, _ := fs.Glob(tfs, "templates/pages/*.tmpl.html")

	byFile := make(map[string]*template.Template)

	// Parse each page template with its associated layouts
	for _, page := range pages {
		name := filepath.Base(page)

		// Create a fresh template instance for each page to avoid cross-contamination
		// This ensures template isolation and prevents naming conflicts
		t := template.New(name).Funcs(funcs)

		// Parse layouts first so pages can reference them with {{template "base" .}}
		// Layouts define the overall page structure and common elements
		t = template.Must(t.ParseFS(tfs, layouts...))

		// Parse the page template which contains the specific content
		// The page template will be executed within the layout context
		t = template.Must(t.ParseFS(tfs, page))

		byFile[name] = t
	}

	return &Renderer{byFile: byFile}
}

// HTML renders a named template to the HTTP response safely.
// It uses a buffer pool for memory efficiency and ensures atomic header writing.
func (r *Renderer) HTML(w http.ResponseWriter, req *http.Request, name string, data any) {
	t, ok := r.byFile[name]
	if !ok {
		http.NotFound(w, req)
		return
	}

	// Use buffer pool for efficient memory management
	// This reduces garbage collection pressure during high-traffic periods
	buf := getBuffer()
	defer putBuffer(buf)

	if err := t.ExecuteTemplate(buf, "base", data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}

	// Set response headers and write content atomically
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}

// Buffer pool for template rendering to reduce memory allocations
// This pool reuses bytes.Buffer instances across multiple template renders
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// getBuffer retrieves a buffer from the pool or creates a new one if needed
func getBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

// putBuffer returns a buffer to the pool after resetting its contents
// This ensures the buffer is clean for the next use
func putBuffer(buf *bytes.Buffer) {
	buf.Reset()
	bufferPool.Put(buf)
}
