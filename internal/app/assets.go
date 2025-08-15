package app

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/fs"
	"strings"
)

// BuildAssetVersions walks the static file system and computes an 8-character SHA-256
// hash for each relevant asset. The hash is used for cache-busting via query strings.
func BuildAssetVersions(static fs.FS) (map[string]string, error) {
	m := make(map[string]string)
	err := fs.WalkDir(static, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		// Only fingerprint common static types
		if !hasAnySuffix(p, ".js", ".css", ".svg", ".json", ".wasm", ".png", ".jpg", ".jpeg", ".gif", "webp", ".avif", ".ico", "woff", ".ttf", ".woff2") {
			return nil
		}
		f, err := static.Open(p)
		if err != nil {
			return err
		}
		h := sha256.New()
		_, _ = io.Copy(h, f)
		_ = f.Close()
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
