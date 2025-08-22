package assets

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"strings"
)

// AssetManager manages static assets with versioning and SRI support.
// It implements the AssetProvider interface.
type AssetManager struct {
	manifest AssetManifest
	logger   *slog.Logger
}

// AssetManifest represents the structure of the manifest file.
// The actual manifest has a "files" wrapper around the asset mappings.
type AssetManifestWrapper struct {
	Files AssetManifest `json:"files"`
}

// New creates a new asset manager instance.
// It loads the manifest from the filesystem and validates assets.
func New(staticFS fs.FS, logger *slog.Logger) (AssetProvider, error) {
	if logger != nil {
		logger.Debug("AssetManager.New called", "staticFS_type", fmt.Sprintf("%T", staticFS))
	}

	// Debug: list available files in the filesystem
	if logger != nil {
		logger.Debug("AssetManager.New listing filesystem contents")
		listFilesystemContents(staticFS, "", logger)
	}

	manifest, err := loadManifest(staticFS)
	if err != nil {
		return nil, fmt.Errorf("failed to load asset manifest: %w", err)
	}

	if logger != nil {
		logger.Debug("AssetManager.New manifest loaded", "manifest_keys", getManifestKeys(manifest))
	}

	return &AssetManager{
		manifest: manifest,
		logger:   logger,
	}, nil
}

// listFilesystemContents recursively lists files in the filesystem for debugging
func listFilesystemContents(fsys fs.FS, path string, logger *slog.Logger) {
	entries, err := fs.ReadDir(fsys, path)
	if err != nil {
		logger.Debug("failed to read directory", "path", path, "error", err)
		return
	}

	for _, entry := range entries {
		fullPath := path
		if fullPath != "" {
			fullPath = fullPath + "/" + entry.Name()
		} else {
			fullPath = entry.Name()
		}

		if entry.IsDir() {
			logger.Debug("directory found", "path", fullPath)
			listFilesystemContents(fsys, fullPath, logger)
		} else {
			logger.Debug("file found", "path", fullPath, "size", entry.Type())
		}
	}
}

// AssetURL returns the versioned URL for an asset.
// It returns the original path if the asset is not found in the manifest.
func (am *AssetManager) AssetURL(path string) string {
	// Try both with and without leading slash
	pathsToTry := []string{path, strings.TrimPrefix(path, "/")}

	if am.logger != nil {
		am.logger.Debug("AssetURL called", "input_path", path, "paths_to_try", pathsToTry, "manifest_keys", getManifestKeys(am.manifest))
	}

	for _, tryPath := range pathsToTry {
		if info, exists := am.manifest[tryPath]; exists {
			// Return versioned path
			if am.logger != nil {
				am.logger.Debug("AssetURL found asset", "path", path, "try_path", tryPath, "returned_path", info.Path)
			}
			return info.Path
		}
	}

	// Return original path if not found in manifest
	am.logger.Warn("asset not found in manifest", "path", path, "paths_tried", pathsToTry, "available_keys", getManifestKeys(am.manifest))
	return path
}

// AssetSRI returns the Subresource Integrity hash for an asset.
// It returns an empty string if the asset is not found in the manifest.
func (am *AssetManager) AssetSRI(path string) string {
	// Try both with and without leading slash
	pathsToTry := []string{path, strings.TrimPrefix(path, "/")}

	if am.logger != nil {
		am.logger.Debug("AssetSRI called", "input_path", path, "paths_to_try", pathsToTry)
	}

	for _, tryPath := range pathsToTry {
		if info, exists := am.manifest[tryPath]; exists {
			if am.logger != nil {
				am.logger.Debug("AssetSRI found asset", "path", path, "try_path", tryPath, "sri", info.SRI)
			}
			return info.SRI
		}
	}

	// Return empty string if not found in manifest
	am.logger.Warn("asset SRI not found in manifest", "path", path, "paths_tried", pathsToTry)
	return ""
}

// GetManifest returns the complete asset manifest.
func (am *AssetManager) GetManifest() AssetManifest {
	return am.manifest
}

// HasAsset returns true if the asset exists in the manifest.
func (am *AssetManager) HasAsset(path string) bool {
	// Try both with and without leading slash
	pathsToTry := []string{path, strings.TrimPrefix(path, "/")}

	for _, tryPath := range pathsToTry {
		if _, exists := am.manifest[tryPath]; exists {
			return true
		}
	}
	return false
}

// GetAssetInfo returns detailed information about an asset.
func (am *AssetManager) GetAssetInfo(path string) (AssetInfo, bool) {
	// Try both with and without leading slash
	pathsToTry := []string{path, strings.TrimPrefix(path, "/")}

	for _, tryPath := range pathsToTry {
		if info, exists := am.manifest[tryPath]; exists {
			return info, true
		}
	}
	return AssetInfo{}, false
}

// loadManifest loads the asset manifest from the filesystem.
// It expects the manifest to be located at "static/dist/js/manifest.json".
func loadManifest(staticFS fs.FS) (AssetManifest, error) {
	// Try different possible paths for the manifest
	possiblePaths := []string{
		"static/dist/js/manifest.json",
		"web/static/dist/js/manifest.json",
		"dist/js/manifest.json",
	}

	var manifestBytes []byte
	var err error
	var usedPath string

	for _, path := range possiblePaths {
		manifestBytes, err = fs.ReadFile(staticFS, path)
		if err == nil {
			usedPath = path
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file from any of the paths %v: %w", possiblePaths, err)
	}

	var wrapper AssetManifestWrapper
	if err := json.Unmarshal(manifestBytes, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse manifest JSON from %s: %w", usedPath, err)
	}

	// Validate manifest structure
	if len(wrapper.Files) == 0 {
		return nil, fmt.Errorf("manifest is empty")
	}

	return wrapper.Files, nil
}

// getManifestKeys returns all available manifest keys for debugging
func getManifestKeys(manifest AssetManifest) []string {
	keys := make([]string, 0, len(manifest))
	for k := range manifest {
		keys = append(keys, k)
	}
	return keys
}
