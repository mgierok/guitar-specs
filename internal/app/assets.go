package app

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/fs"
	"strings"
)

const (
	// MaxFileSize limits the size of files that can be processed for asset versioning
	// This prevents memory exhaustion from extremely large files
	MaxFileSize = 10 * 1024 * 1024 // 10MB
)

// BuildAssetVersions walks the static file system and computes an 8-character SHA-256
// hash for each relevant asset. The hash is used for cache-busting via query strings.
// This function requires all assets to be processed successfully for application stability.
// Large files (>10MB) are skipped to prevent memory exhaustion.
func BuildAssetVersions(static fs.FS) (map[string]string, error) {
	m := make(map[string]string)

	err := fs.WalkDir(static, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // Fail fast on any filesystem error
		}

		if d.IsDir() {
			return nil
		}

		// Only fingerprint common static types
		if !hasAnySuffix(p, ".js", ".css", ".svg", ".json", ".wasm", ".png", ".jpg", ".jpeg", ".gif", "webp", ".avif", ".ico", "woff", ".ttf", ".woff2") {
			return nil
		}

		// Check file size before processing
		info, err := d.Info()
		if err != nil {
			return err // Fail fast on file info errors
		}

		if info.Size() > MaxFileSize {
			// Skip large files to prevent memory exhaustion
			// These files will still be served but without versioning
			return nil
		}

		f, err := static.Open(p)
		if err != nil {
			return err // Fail fast on file open errors
		}
		defer f.Close()

		// Use streaming hash to avoid loading entire file into memory
		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			return err // Fail fast on file read errors
		}

		sum := hex.EncodeToString(h.Sum(nil))[:8]
		// Normalise to URL path under /static
		urlPath := "/static/" + strings.TrimPrefix(p, "./")
		urlPath = strings.ReplaceAll(urlPath, "\\", "/")
		m[urlPath] = sum
		return nil
	})

	return m, err
}

// hasAnySuffix checks if s has any of the given suffixes (case-insensitive).
func hasAnySuffix(s string, suff ...string) bool {
	for _, x := range suff {
		if strings.HasSuffix(strings.ToLower(s), x) {
			return true
		}
	}
	return false
}
