package app

import (
	"encoding/json"
	"io/fs"
	"log/slog"
	"strings"
)

// AssetManifest represents the structure of the manifest.json file
type AssetManifest struct {
	Files map[string]AssetInfo `json:"files"`
}

// AssetInfo contains information about a hashed asset
type AssetInfo struct {
	Hashed string `json:"hashed"`
	SRI    string `json:"sri"`
	Hash   string `json:"hash"`
	Path   string `json:"path"`
}

// AssetManager handles asset versioning and SRI from build-time manifest
type AssetManager struct {
	manifest  *AssetManifest
	enableSRI bool // Flag to enable/disable SRI (useful with Cloudflare compression)
	logger    *slog.Logger
}

// NewAssetManager creates a new asset manager from manifest file
func NewAssetManager(static fs.FS, logger *slog.Logger) (*AssetManager, error) {
	manifestData, err := fs.ReadFile(static, "static/dist/js/manifest.json")
	if err != nil {
		return nil, err
	}

	if logger != nil {
		logger.Debug("read manifest.json", "bytes", len(manifestData))
	}

	var manifest AssetManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return nil, err
	}

	if logger != nil {
		logger.Debug("parsed manifest", "files", len(manifest.Files))
	}

	// SRI can be disabled if Cloudflare compresses assets
	enableSRI := true // Can be controlled via environment variable

	return &AssetManager{manifest: &manifest, enableSRI: enableSRI, logger: logger}, nil
}

// AssetURL returns the hashed URL for an asset
func (am *AssetManager) AssetURL(assetPath string) string {
	if am.manifest == nil {
		panic("asset manager not initialized - manifest required")
	}

	if am.logger != nil {
		am.logger.Debug("looking for asset", "asset", assetPath)
	}

	// Look for the asset in manifest
	for originalPath, info := range am.manifest.Files {
		if strings.HasSuffix(originalPath, assetPath) {
			if am.logger != nil {
				am.logger.Debug("resolved asset", "asset", assetPath, "path", info.Path)
			}
			return info.Path
		}
	}

	// Panic if asset not found in manifest
	panic("asset not found in manifest: " + assetPath)
}

// AssetSRI returns the SRI hash for an asset
func (am *AssetManager) AssetSRI(assetPath string) string {
	if am.manifest == nil {
		panic("asset manager not initialized - manifest required")
	}

	// Return empty string if SRI is disabled (e.g., Cloudflare compression)
	if !am.enableSRI {
		return ""
	}

	// Look for the asset in manifest
	for originalPath, info := range am.manifest.Files {
		if strings.HasSuffix(originalPath, assetPath) {
			return info.SRI
		}
	}

	// Panic if asset not found in manifest
	panic("asset not found in manifest: " + assetPath)
}

// AssetWithSRI returns the asset URL with SRI attribute if available
func (am *AssetManager) AssetWithSRI(assetPath string) (string, string) {
	url := am.AssetURL(assetPath)
	sri := am.AssetSRI(assetPath)
	return url, sri
}

// HasAsset checks if an asset exists in the manifest
func (am *AssetManager) HasAsset(assetPath string) bool {
	if am.manifest == nil {
		return false
	}

	for originalPath := range am.manifest.Files {
		if strings.HasSuffix(originalPath, assetPath) {
			return true
		}
	}
	return false
}
