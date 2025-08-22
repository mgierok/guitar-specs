package assets

// AssetProvider defines the interface for static asset management.
// This interface allows for dependency inversion and easier testing.
type AssetProvider interface {
	// AssetURL returns the versioned URL for an asset
	// Example: "/static/css/main.css" -> "/static/css/main.abc123.css"
	AssetURL(path string) string

	// AssetSRI returns the Subresource Integrity hash for an asset
	// Example: "sha384-abc123def456..."
	AssetSRI(path string) string

	// GetManifest returns the complete asset manifest
	GetManifest() AssetManifest

	// HasAsset returns true if the asset exists in the manifest
	HasAsset(path string) bool

	// GetAssetInfo returns detailed information about an asset
	GetAssetInfo(path string) (AssetInfo, bool)
}

// AssetManifest represents the complete asset manifest structure
type AssetManifest map[string]AssetInfo

// AssetInfo holds information about a single asset
type AssetInfo struct {
	// Original path of the asset
	Path string `json:"path"`

	// Versioned filename with hash
	Filename string `json:"filename"`

	// Subresource Integrity hash
	SRI string `json:"sri"`

	// File size in bytes
	Size int64 `json:"size"`

	// MIME type
	ContentType string `json:"content_type"`
}
