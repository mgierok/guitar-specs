package render

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"guitar-specs/internal/assets"
)

// TemplateRenderer manages HTML template rendering with asset helper functions.
// It implements the Renderer interface.
type TemplateRenderer struct {
	templates map[string]*template.Template
	funcs     template.FuncMap
	env       string
	logger    *slog.Logger
	mu        sync.RWMutex
}

// New creates a new template renderer instance.
// It parses all templates from the filesystem and sets up helper functions.
func New(templatesFS fs.FS, assetProvider assets.AssetProvider, env string, logger *slog.Logger) (Renderer, error) {
	// Create template function map with asset helpers
	funcs := template.FuncMap{
		"asset": assetProvider.AssetURL,
		"sri":   assetProvider.AssetSRI,
	}

	if logger != nil {
		logger.Debug("Renderer.New creating function map", "funcs_count", len(funcs), "funcs", getFuncNames(funcs))
	}

	renderer := &TemplateRenderer{
		templates: make(map[string]*template.Template),
		funcs:     funcs,
		env:       env,
		logger:    logger,
	}

	// Parse all templates
	if err := renderer.parseTemplates(templatesFS); err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return renderer, nil
}

// getFuncNames returns function names for debugging
func getFuncNames(funcs template.FuncMap) []string {
	names := make([]string, 0, len(funcs))
	for name := range funcs {
		names = append(names, name)
	}
	return names
}

// Render renders a template with the given data and writes to the writer.
func (r *TemplateRenderer) Render(w io.Writer, templateName string, data interface{}) error {
	r.mu.RLock()
	tmpl, exists := r.templates[templateName]
	r.mu.RUnlock()

	if r.logger != nil {
		r.logger.Debug("rendering template", "name", templateName, "exists", exists, "available_templates", r.getTemplateNames())
	}

	if !exists {
		return fmt.Errorf("template '%s' not found", templateName)
	}

	// Prepare template data with common functions
	templateData := r.prepareTemplateData(data)

	// Execute template
	if err := tmpl.Execute(w, templateData); err != nil {
		return fmt.Errorf("failed to execute template '%s': %w", templateName, err)
	}

	return nil
}

// RenderWithRequest renders a template with request context for CSP nonce.
func (r *TemplateRenderer) RenderWithRequest(w io.Writer, templateName string, req *http.Request, data interface{}) error {
	r.mu.RLock()
	tmpl, exists := r.templates[templateName]
	r.mu.RUnlock()

	if r.logger != nil {
		r.logger.Debug("rendering template with request", "name", templateName, "exists", exists, "available_templates", r.getTemplateNames())
	}

	if !exists {
		return fmt.Errorf("template '%s' not found", templateName)
	}

	// Prepare template data with common functions and request context
	templateData := r.prepareTemplateDataWithRequest(data, req)

	// Execute template
	if err := tmpl.Execute(w, templateData); err != nil {
		return fmt.Errorf("failed to execute template '%s': %w", templateName, err)
	}

	return nil
}

// RenderString renders a template and returns the result as a string.
func (r *TemplateRenderer) RenderString(templateName string, data interface{}) (string, error) {
	var buf bytes.Buffer

	if err := r.Render(&buf, templateName, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// GetTemplate returns a specific template by name.
func (r *TemplateRenderer) GetTemplate(name string) (*template.Template, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tmpl, exists := r.templates[name]
	if !exists {
		return nil, fmt.Errorf("template '%s' not found", name)
	}

	return tmpl, nil
}

// GetTemplates returns all available templates.
func (r *TemplateRenderer) GetTemplates() map[string]*template.Template {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external modification
	templates := make(map[string]*template.Template)
	for name, tmpl := range r.templates {
		templates[name] = tmpl
	}

	return templates
}

// AddTemplate adds a new template to the renderer.
func (r *TemplateRenderer) AddTemplate(name string, tmpl *template.Template) error {
	if name == "" {
		return fmt.Errorf("template name cannot be empty")
	}

	if tmpl == nil {
		return fmt.Errorf("template cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.templates[name] = tmpl

	if r.logger != nil {
		r.logger.Debug("added template", "name", name)
	}

	return nil
}

// HasTemplate returns true if the template exists.
func (r *TemplateRenderer) HasTemplate(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.templates[name]
	return exists
}

// getTemplateNames returns a list of available template names for debugging.
func (r *TemplateRenderer) getTemplateNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.templates))
	for name := range r.templates {
		names = append(names, name)
	}
	return names
}

// parseTemplates discovers and parses all templates from the filesystem.
func (r *TemplateRenderer) parseTemplates(templatesFS fs.FS) error {
	// Discover layout templates
	layouts, err := fs.Glob(templatesFS, "templates/layouts/*.tmpl.html")
	if err != nil {
		return fmt.Errorf("failed to discover layout templates: %w", err)
	}

	if r.logger != nil {
		r.logger.Info("discovered layout templates", "count", len(layouts), "layouts", layouts)
	}

	// Discover page templates
	pages, err := fs.Glob(templatesFS, "templates/pages/*.tmpl.html")
	if err != nil {
		return fmt.Errorf("failed to discover page templates: %w", err)
	}

	if r.logger != nil {
		r.logger.Info("discovered page templates", "count", len(pages), "pages", pages)
	}

	if len(pages) == 0 {
		return fmt.Errorf("no page templates found")
	}

	// Parse each page template with its associated layouts
	for _, page := range pages {
		name := filepath.Base(page)
		shortName := strings.TrimSuffix(name, ".tmpl.html")

		// Create new template with helper functions FIRST
		tmpl := template.New(name).Funcs(r.funcs)

		// Parse layouts first
		for _, layout := range layouts {
			tmpl = template.Must(tmpl.ParseFS(templatesFS, layout))
		}

		// Parse page content
		tmpl = template.Must(tmpl.ParseFS(templatesFS, page))

		// Store with both full name and short name
		r.templates[name] = tmpl
		r.templates[shortName] = tmpl

		if r.logger != nil {
			r.logger.Debug("parsed template", "name", name, "shortName", shortName, "has_funcs", len(r.funcs))
		}
	}

	return nil
}

// prepareTemplateData prepares template data with common functions and environment info.
func (r *TemplateRenderer) prepareTemplateData(data interface{}) interface{} {
	// If data is already TemplateData, return as is
	if td, ok := data.(TemplateData); ok {
		return td
	}

	// If data is map, wrap it in TemplateData structure
	if m, ok := data.(map[string]interface{}); ok {
		return TemplateData{
			Page: m,
			Common: CommonData{
				Environment: r.env,
			},
		}
	}

	// Create new TemplateData with common info
	return TemplateData{
		Page: data,
		Common: CommonData{
			Environment: r.env,
		},
	}
}

// prepareTemplateDataWithRequest prepares template data with request context for CSP nonce.
func (r *TemplateRenderer) prepareTemplateDataWithRequest(data interface{}, req *http.Request) interface{} {
	// If data is already TemplateData, return as is
	if td, ok := data.(TemplateData); ok {
		// Add CSP nonce if available
		if nonce, ok := req.Context().Value("cspNonce").(string); ok {
			td.Common.CSPNonce = nonce
		}
		return td
	}

	// If data is map, wrap it in TemplateData structure
	if m, ok := data.(map[string]interface{}); ok {
		common := CommonData{
			Environment: r.env,
		}

		// Add CSP nonce if available
		if nonce, ok := req.Context().Value("cspNonce").(string); ok {
			common.CSPNonce = nonce
		}

		return TemplateData{
			Page:   m,
			Common: common,
		}
	}

	// Create new TemplateData with common info
	common := CommonData{
		Environment: r.env,
	}

	// Add CSP nonce if available
	if nonce, ok := req.Context().Value("cspNonce").(string); ok {
		common.CSPNonce = nonce
	}

	return TemplateData{
		Page:   data,
		Common: common,
	}
}
