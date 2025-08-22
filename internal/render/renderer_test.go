package render

import (
	"bytes"
	"html/template"
	"log/slog"
	"os"
	"testing"
	"testing/fstest"

	"guitar-specs/internal/assets"
)

// MockAssetProvider implements assets.AssetProvider for testing
type MockAssetProvider struct {
	assetURLs map[string]string
	assetSRIs map[string]string
}

func (m *MockAssetProvider) AssetURL(path string) string {
	if url, exists := m.assetURLs[path]; exists {
		return url
	}
	return path
}

func (m *MockAssetProvider) AssetSRI(path string) string {
	if sri, exists := m.assetSRIs[path]; exists {
		return sri
	}
	return ""
}

func (m *MockAssetProvider) GetManifest() assets.AssetManifest {
	return make(assets.AssetManifest)
}

func (m *MockAssetProvider) HasAsset(path string) bool {
	_, exists := m.assetURLs[path]
	return exists
}

func (m *MockAssetProvider) GetAssetInfo(path string) (assets.AssetInfo, bool) {
	return assets.AssetInfo{}, false
}

func TestNew(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	
	// Create mock asset provider
	mockAssets := &MockAssetProvider{
		assetURLs: map[string]string{
			"/static/css/main.css": "/static/css/main.abc123.css",
			"/static/js/app.js":    "/static/js/app.def456.js",
		},
		assetSRIs: map[string]string{
			"/static/css/main.css": "sha384-abc123",
			"/static/js/app.js":    "sha384-def456",
		},
	}
	
	// Create mock templates filesystem
	mockFS := fstest.MapFS{
		"templates/layouts/base.tmpl.html": &fstest.MapFile{
			Data: []byte(`<!DOCTYPE html>
<html>
<head>
    <title>{{.Page.Title}}</title>
    <link rel="stylesheet" href="{{asset "/static/css/main.css"}}" integrity="{{sri "/static/css/main.css"}}">
</head>
<body>
    {{template "content" .}}
    <script src="{{asset "/static/js/app.js"}}" integrity="{{sri "/static/js/app.js"}}"></script>
</body>
</html>`),
		},
		"templates/pages/home.tmpl.html": &fstest.MapFile{
			Data: []byte(`{{define "content"}}
<h1>{{.Page.Title}}</h1>
<p>{{.Page.Content}}</p>
{{end}}`),
		},
	}
	
	// Test with valid templates
	renderer, err := New(mockFS, mockAssets, "development", logger)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Test that it implements the interface
	var _ Renderer = renderer
	
	// Test that templates were loaded
	if !renderer.HasTemplate("home.tmpl.html") {
		t.Error("Expected home.tmpl.html template to exist")
	}
	
	if !renderer.HasTemplate("home") {
		t.Error("Expected 'home' template to exist")
	}
}

func TestNewWithInvalidTemplates(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	
	// Create mock asset provider
	mockAssets := &MockAssetProvider{
		assetURLs: make(map[string]string),
		assetSRIs: make(map[string]string),
	}
	
	// Test with empty filesystem
	emptyFS := fstest.MapFS{}
	
	_, err := New(emptyFS, mockAssets, "development", logger)
	if err == nil {
		t.Error("Expected error with empty filesystem, got nil")
	}
}

func TestTemplateRenderer_Render(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	
	// Create mock asset provider
	mockAssets := &MockAssetProvider{
		assetURLs: map[string]string{
			"/static/css/main.css": "/static/css/main.abc123.css",
		},
		assetSRIs: map[string]string{
			"/static/css/main.css": "sha384-abc123",
		},
	}
	
	// Create simple template
	mockFS := fstest.MapFS{
		"templates/pages/simple.tmpl.html": &fstest.MapFile{
			Data: []byte(`<h1>{{.Page.Title}}</h1>
<p>{{.Page.Content}}</p>
<link href="{{asset "/static/css/main.css"}}" integrity="{{sri "/static/css/main.css"}}">`),
		},
	}
	
	renderer, err := New(mockFS, mockAssets, "development", logger)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Test data
	data := map[string]interface{}{
		"Title":   "Test Page",
		"Content": "This is a test",
	}
	
	// Test Render to writer
	var buf bytes.Buffer
	err = renderer.Render(&buf, "simple", data)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	result := buf.String()
	if !contains(result, "Test Page") {
		t.Errorf("Expected 'Test Page' in output, got: %s", result)
	}
	
	if !contains(result, "This is a test") {
		t.Errorf("Expected 'This is a test' in output, got: %s", result)
	}
	
	if !contains(result, "/static/css/main.abc123.css") {
		t.Errorf("Expected asset URL in output, got: %s", result)
	}
	
	if !contains(result, "sha384-abc123") {
		t.Errorf("Expected SRI hash in output, got: %s", result)
	}
}

func TestTemplateRenderer_RenderString(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	
	// Create mock asset provider
	mockAssets := &MockAssetProvider{
		assetURLs: map[string]string{
			"/static/css/main.css": "/static/css/main.abc123.css",
		},
		assetSRIs: map[string]string{
			"/static/css/main.css": "sha384-abc123",
		},
	}
	
	// Create simple template
	mockFS := fstest.MapFS{
		"templates/pages/string.tmpl.html": &fstest.MapFile{
			Data: []byte(`<h1>{{.Page.Title}}</h1>`),
		},
	}
	
	renderer, err := New(mockFS, mockAssets, "development", logger)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Test data
	data := map[string]interface{}{
		"Title": "String Test",
	}
	
	// Test RenderString
	result, err := renderer.RenderString("string", data)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if !contains(result, "String Test") {
		t.Errorf("Expected 'String Test' in output, got: %s", result)
	}
}

func TestTemplateRenderer_GetTemplate(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	
	// Create mock asset provider
	mockAssets := &MockAssetProvider{
		assetURLs: make(map[string]string),
		assetSRIs: make(map[string]string),
	}
	
	// Create template
	mockFS := fstest.MapFS{
		"templates/pages/test.tmpl.html": &fstest.MapFile{
			Data: []byte(`<h1>Test</h1>`),
		},
	}
	
	renderer, err := New(mockFS, mockAssets, "development", logger)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Test GetTemplate
	tmpl, err := renderer.GetTemplate("test")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if tmpl == nil {
		t.Error("Expected non-nil template")
	}
	
	// Test GetTemplate with non-existent template
	_, err = renderer.GetTemplate("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent template")
	}
}

func TestTemplateRenderer_GetTemplates(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	
	// Create mock asset provider
	mockAssets := &MockAssetProvider{
		assetURLs: make(map[string]string),
		assetSRIs: make(map[string]string),
	}
	
	// Create templates
	mockFS := fstest.MapFS{
		"templates/pages/page1.tmpl.html": &fstest.MapFile{
			Data: []byte(`<h1>Page 1</h1>`),
		},
		"templates/pages/page2.tmpl.html": &fstest.MapFile{
			Data: []byte(`<h1>Page 2</h1>`),
		},
	}
	
	renderer, err := New(mockFS, mockAssets, "development", logger)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Test GetTemplates
	templates := renderer.GetTemplates()
	
	// Should have both full names and short names
	expectedCount := 4 // page1.tmpl.html, page1, page2.tmpl.html, page2
	if len(templates) != expectedCount {
		t.Errorf("Expected %d templates, got %d", expectedCount, len(templates))
	}
	
	// Check specific templates
	if _, exists := templates["page1"]; !exists {
		t.Error("Expected 'page1' template to exist")
	}
	
	if _, exists := templates["page2"]; !exists {
		t.Error("Expected 'page2' template to exist")
	}
}

func TestTemplateRenderer_AddTemplate(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	
	// Create mock asset provider
	mockAssets := &MockAssetProvider{
		assetURLs: make(map[string]string),
		assetSRIs: make(map[string]string),
	}
	
	// Create empty filesystem
	emptyFS := fstest.MapFS{}
	
	renderer, err := New(emptyFS, mockAssets, "development", logger)
	if err == nil {
		t.Error("Expected error with empty filesystem")
		return
	}
	
	// Create a simple renderer manually for testing AddTemplate
	renderer = &TemplateRenderer{
		templates: make(map[string]*template.Template),
		funcs:     template.FuncMap{},
		env:       "test",
		logger:    logger,
	}
	
	// Create a test template
	testTmpl := template.Must(template.New("test").Parse("<h1>Test</h1>"))
	
	// Test AddTemplate
	err = renderer.AddTemplate("test", testTmpl)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Verify template was added
	if !renderer.HasTemplate("test") {
		t.Error("Expected 'test' template to exist after adding")
	}
	
	// Test AddTemplate with empty name
	err = renderer.AddTemplate("", testTmpl)
	if err == nil {
		t.Error("Expected error for empty template name")
	}
	
	// Test AddTemplate with nil template
	err = renderer.AddTemplate("nil", nil)
	if err == nil {
		t.Error("Expected error for nil template")
	}
}

func TestTemplateRenderer_HasTemplate(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	
	// Create mock asset provider
	mockAssets := &MockAssetProvider{
		assetURLs: make(map[string]string),
		assetSRIs: make(map[string]string),
	}
	
	// Create template
	mockFS := fstest.MapFS{
		"templates/pages/exists.tmpl.html": &fstest.MapFile{
			Data: []byte(`<h1>Exists</h1>`),
		},
	}
	
	renderer, err := New(mockFS, mockAssets, "development", logger)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Test HasTemplate with existing template
	if !renderer.HasTemplate("exists") {
		t.Error("Expected 'exists' template to exist")
	}
	
	// Test HasTemplate with non-existent template
	if renderer.HasTemplate("nonexistent") {
		t.Error("Expected 'nonexistent' template to not exist")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
