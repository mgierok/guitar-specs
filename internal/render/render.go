package render

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"guitar-specs/internal/http/middleware"
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
	layouts, err := fs.Glob(tfs, "templates/layouts/*.tmpl.html")
	if err != nil {
		panic(fmt.Errorf("failed to discover layout templates: %w", err))
	}
	pages, err := fs.Glob(tfs, "templates/pages/*.tmpl.html")
	if err != nil {
		panic(fmt.Errorf("failed to discover page templates: %w", err))
	}

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

		// Store template with both the filename and "content" as keys for flexibility
		byFile[name] = t
		byFile[strings.TrimSuffix(name, ".tmpl.html")] = t
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

	// If CSP nonce is present in context, expose it to templates as .cspNonce
	if nonce, ok := middleware.CSPNonceFromContext(req.Context()); ok {
		if m, ok2 := data.(map[string]any); ok2 {
			m["cspNonce"] = nonce
		}
	}

	// Execute the template - it should contain both "base" and "content" blocks
	if err := t.ExecuteTemplate(buf, "base", data); err != nil {
		// Log detailed template error for debugging
		fmt.Printf("Template execution error for %s: %v\n", name, err)
		fmt.Printf("Available templates: %v\n", t.DefinedTemplates())
		fmt.Printf("Template name: %s\n", name)
		
		// Return detailed error to help with debugging
		http.Error(w, fmt.Sprintf("template error: %v", err), http.StatusInternalServerError)
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
