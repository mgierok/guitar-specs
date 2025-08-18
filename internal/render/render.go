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

type Renderer struct {
	byFile map[string]*template.Template
	env    string
}

func New(tfs embed.FS, funcs template.FuncMap, env string) *Renderer {
	// Discover and parse all templates
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
		t := template.New(name).Funcs(funcs)
		
		// Parse layouts first, then page content
		t = template.Must(t.ParseFS(tfs, layouts...))
		t = template.Must(t.ParseFS(tfs, page))

		// Store with both full name and short name for flexibility
		byFile[name] = t
		byFile[strings.TrimSuffix(name, ".tmpl.html")] = t
		
		if env == "development" {
			fmt.Printf("✓ Parsed template: %s\n", name)
		}
	}

	return &Renderer{byFile: byFile, env: env}
}

func (r *Renderer) HTML(w http.ResponseWriter, req *http.Request, name string, data any) {
	t, ok := r.byFile[name]
	if !ok {
		r.handleTemplateNotFound(w, name)
		return
	}

	buf := getBuffer()
	defer putBuffer(buf)

	// Inject CSP nonce if available
	if nonce, ok := middleware.CSPNonceFromContext(req.Context()); ok {
		if m, ok2 := data.(map[string]any); ok2 {
			m["cspNonce"] = nonce
		}
	}

	// Execute template with "base" block
	if err := t.ExecuteTemplate(buf, "base", data); err != nil {
		r.handleTemplateError(w, name, t, err)
		return
	}

	// Write response
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}

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

// listAvailableTemplates returns a formatted list of available templates for debugging
func (r *Renderer) listAvailableTemplates() string {
	if len(r.byFile) == 0 {
		return "No templates available"
	}
	
	var templates []string
	for name := range r.byFile {
		templates = append(templates, name)
	}
	
	return strings.Join(templates, "\n- ")
}

// handleTemplateNotFound handles cases where a template is not found
func (r *Renderer) handleTemplateNotFound(w http.ResponseWriter, name string) {
	fmt.Printf("🚨 Template not found: %s\n", name)
	fmt.Printf("📁 Available templates: %v\n", r.listAvailableTemplates())
	
	if r.env == "development" {
		http.Error(w, fmt.Sprintf(`Template Not Found: %s

Available Templates:
%s

This detailed error is shown in development mode only.`, 
			name, r.listAvailableTemplates()), http.StatusNotFound)
	} else {
		http.Error(w, "Page Not Found", http.StatusNotFound)
	}
}

// handleTemplateError handles template execution errors
func (r *Renderer) handleTemplateError(w http.ResponseWriter, name string, t *template.Template, err error) {
	fmt.Printf("🚨 Template execution error for %s: %v\n", name, err)
	fmt.Printf("📁 Available templates: %s\n", t.DefinedTemplates())
	
	if r.env == "development" {
		errorMsg := fmt.Sprintf(`Template Error: %v

Template: %s
Available Templates: %s

This detailed error is shown in development mode only.`, 
			err, name, t.DefinedTemplates())
		http.Error(w, errorMsg, http.StatusInternalServerError)
	} else {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
