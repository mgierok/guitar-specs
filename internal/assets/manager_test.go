package assets

import (
	"log/slog"
	"os"
	"testing"
	"testing/fstest"
)

func TestNew(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	// Create mock filesystem with valid manifest
	mockFS := fstest.MapFS{
		"static/dist/js/manifest.json": &fstest.MapFile{
			Data: []byte(`{
					"files": {
						"static/css/main.css": {
							"path": "/static/css/main.abc123.css",
							"filename": "static/css/main.abc123.css",
							"sri": "sha384-abc123def456ghi789",
							"size": 1024,
							"content_type": "text/css"
						},
						"static/js/app.js": {
							"path": "/static/js/app.def456.js",
							"filename": "static/js/app.def456.js",
							"sri": "sha384-def456ghi789abc123",
							"size": 2048,
							"content_type": "application/javascript"
						}
					}
				}`),
		},
	}

	// Test with valid filesystem
	assetManager, err := New(mockFS, logger)
	if err != nil {
		t.Fatalf("Expected no error with valid manifest, got %v", err)
	}

	// Test that it implements the interface
	var _ AssetProvider = assetManager

	// Test that manifest was loaded
	manifest := assetManager.GetManifest()
	if len(manifest) != 2 {
		t.Errorf("Expected manifest with 2 items, got %d", len(manifest))
	}

	// Test specific assets from manifest
	if !assetManager.HasAsset("static/css/main.css") {
		t.Error("Expected static/css/main.css to exist in manifest")
	}

	if !assetManager.HasAsset("static/js/app.js") {
		t.Error("Expected static/js/app.js to exist in manifest")
	}
}

func TestNewWithManifestData(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	// Create mock filesystem with manifest data
	mockFS := fstest.MapFS{
		"static/dist/js/manifest.json": &fstest.MapFile{
			Data: []byte(`{
				"files": {
					"static/css/main.css": {
						"path": "/static/css/main.abc123.css",
						"filename": "static/css/main.abc123.css",
						"sri": "sha384-abc123def456ghi789",
						"size": 1024,
						"content_type": "text/css"
					},
					"static/js/app.js": {
						"path": "/static/js/app.def456.js",
						"filename": "static/js/app.def456.js",
						"sri": "sha384-def456ghi789abc123",
						"size": 2048,
						"content_type": "application/javascript"
					}
				}
			}`),
		},
	}

	assetManager, err := New(mockFS, logger)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test AssetURL with manifest data
	url := assetManager.AssetURL("/static/css/main.css")
	expectedURL := "/static/css/main.abc123.css"
	if url != expectedURL {
		t.Errorf("Expected URL %s, got %s", expectedURL, url)
	}

	// Test AssetSRI with manifest data
	sri := assetManager.AssetSRI("/static/css/main.css")
	expectedSRI := "sha384-abc123def456ghi789"
	if sri != expectedSRI {
		t.Errorf("Expected SRI %s, got %s", expectedSRI, sri)
	}

	// Test GetAssetInfo with manifest data
	info, exists := assetManager.GetAssetInfo("static/js/app.js")
	if !exists {
		t.Error("Expected static/js/app.js to exist")
	}

	if info.Filename != "static/js/app.def456.js" {
		t.Errorf("Expected filename static/js/app.def456.js, got %s", info.Filename)
	}

	if info.SRI != "sha384-def456ghi789abc123" {
		t.Errorf("Expected SRI sha384-def456ghi789abc123, got %s", info.SRI)
	}

	if info.Size != 2048 {
		t.Errorf("Expected size 2048, got %d", info.Size)
	}

	if info.ContentType != "application/javascript" {
		t.Errorf("Expected content type application/javascript, got %s", info.ContentType)
	}
}

func TestNewWithInvalidFilesystem(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	// Create empty mock filesystem (no manifest)
	emptyFS := fstest.MapFS{}

	// Test with invalid filesystem (no manifest)
	_, err := New(emptyFS, logger)
	if err == nil {
		t.Error("Expected error when manifest doesn't exist, got nil")
	}
}

func TestNewWithInvalidJSON(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	// Create mock filesystem with invalid JSON
	invalidFS := fstest.MapFS{
		"static/dist/js/manifest.json": &fstest.MapFile{
			Data: []byte(`{invalid json`),
		},
	}

	// Test with invalid JSON manifest
	_, err := New(invalidFS, logger)
	if err == nil {
		t.Error("Expected error when manifest JSON is invalid, got nil")
	}
}

func TestNewWithEmptyManifest(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	// Create mock filesystem with empty manifest
	emptyManifestFS := fstest.MapFS{
		"static/dist/js/manifest.json": &fstest.MapFile{
			Data: []byte(`{}`),
		},
	}

	// Test with empty manifest
	_, err := New(emptyManifestFS, logger)
	if err == nil {
		t.Error("Expected error when manifest is empty, got nil")
	}
}

func TestAssetManager_AssetURL(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	// Create a mock asset manager with test data
	am := &AssetManager{
		manifest: AssetManifest{
			"static/css/main.css": AssetInfo{
				Path:        "/static/css/main.abc123.css",
				Filename:    "static/css/main.abc123.css",
				SRI:         "sha384-abc123",
				Size:        1024,
				ContentType: "text/css",
			},
			"static/js/app.js": AssetInfo{
				Path:        "/static/js/app.def456.js",
				Filename:    "static/js/app.def456.js",
				SRI:         "sha384-def456",
				Size:        2048,
				ContentType: "application/javascript",
			},
		},
		logger: logger,
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "CSS file with leading slash",
			input:    "/static/css/main.css",
			expected: "/static/css/main.abc123.css",
		},
		{
			name:     "CSS file without leading slash",
			input:    "static/css/main.css",
			expected: "/static/css/main.abc123.css",
		},
		{
			name:     "JS file",
			input:    "/static/js/app.js",
			expected: "/static/js/app.def456.js",
		},
		{
			name:     "Non-existent file",
			input:    "/static/css/notfound.css",
			expected: "/static/css/notfound.css",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := am.AssetURL(tt.input)
			if result != tt.expected {
				t.Errorf("AssetURL(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestAssetManager_AssetSRI(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	am := &AssetManager{
		manifest: AssetManifest{
			"static/css/main.css": AssetInfo{
				Path:        "/static/css/main.abc123.css",
				Filename:    "static/css/main.abc123.css",
				SRI:         "sha384-abc123",
				Size:        1024,
				ContentType: "text/css",
			},
		},
		logger: logger,
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Existing CSS file",
			input:    "/static/css/main.css",
			expected: "sha384-abc123",
		},
		{
			name:     "Non-existent file",
			input:    "/static/css/notfound.css",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := am.AssetSRI(tt.input)
			if result != tt.expected {
				t.Errorf("AssetSRI(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestAssetManager_HasAsset(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	am := &AssetManager{
		manifest: AssetManifest{
			"static/css/main.css": AssetInfo{
				Path:     "/static/css/main.abc123.css",
				Filename: "static/css/main.abc123.css",
			},
		},
		logger: logger,
	}

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Existing asset",
			input:    "/static/css/main.css",
			expected: true,
		},
		{
			name:     "Non-existent asset",
			input:    "/static/css/notfound.css",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := am.HasAsset(tt.input)
			if result != tt.expected {
				t.Errorf("HasAsset(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestAssetManager_GetAssetInfo(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	expectedInfo := AssetInfo{
		Path:        "/static/css/main.abc123.css",
		Filename:    "static/css/main.abc123.css",
		SRI:         "sha384-abc123",
		Size:        1024,
		ContentType: "text/css",
	}

	am := &AssetManager{
		manifest: AssetManifest{
			"static/css/main.css": expectedInfo,
		},
		logger: logger,
	}

	// Test existing asset
	info, exists := am.GetAssetInfo("/static/css/main.css")
	if !exists {
		t.Error("Expected asset to exist")
	}
	if info != expectedInfo {
		t.Errorf("GetAssetInfo() = %+v, want %+v", info, expectedInfo)
	}

	// Test non-existent asset
	_, exists = am.GetAssetInfo("/static/css/notfound.css")
	if exists {
		t.Error("Expected asset to not exist")
	}
}

func TestAssetManager_GetManifest(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	expectedManifest := AssetManifest{
		"static/css/main.css": AssetInfo{
			Path:     "/static/css/main.abc123.css",
			Filename: "static/css/main.abc123.css",
		},
	}

	am := &AssetManager{
		manifest: expectedManifest,
		logger:   logger,
	}

	manifest := am.GetManifest()
	if len(manifest) != len(expectedManifest) {
		t.Errorf("GetManifest() returned %d items, want %d", len(manifest), len(expectedManifest))
	}
}
