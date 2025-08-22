package render

import (
	"html/template"
	"io"
	"net/http"
)

// Renderer defines the interface for HTML template rendering.
// This interface allows for dependency inversion and easier testing.
type Renderer interface {
	// Render renders a template with the given data and writes to the writer
	Render(w io.Writer, templateName string, data interface{}) error

	// RenderWithRequest renders a template with request context for CSP nonce
	RenderWithRequest(w io.Writer, templateName string, req *http.Request, data interface{}) error

	// RenderString renders a template and returns the result as a string
	RenderString(templateName string, data interface{}) (string, error)

	// GetTemplate returns a specific template by name
	GetTemplate(name string) (*template.Template, error)

	// GetTemplates returns all available templates
	GetTemplates() map[string]*template.Template

	// AddTemplate adds a new template to the renderer
	AddTemplate(name string, tmpl *template.Template) error

	// HasTemplate returns true if the template exists
	HasTemplate(name string) bool
}

// TemplateData represents common data passed to all templates
type TemplateData struct {
	// Page-specific data
	Page interface{}

	// Common data for all pages
	Common CommonData
}

// CommonData represents data shared across all templates
type CommonData struct {
	// Environment (development, production, etc.)
	Environment string

	// Asset helper functions
	AssetURL func(string) string
	AssetSRI func(string) string

	// CSP nonce for security
	CSPNonce string

	// Other common data can be added here
	Version   string
	BuildTime string
}
