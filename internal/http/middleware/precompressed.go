package middleware

import (
	"io/fs"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// PrecompressedFileServer serves files, preferring pre-compressed variants (.br → .gz) when the client supports them.
// Comments in British English as requested.
func PrecompressedFileServer(root fs.FS) http.Handler {
	base := http.FileServer(http.FS(root))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only optimised for GET/HEAD; fall back for others
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			base.ServeHTTP(w, r)
			return
		}

		// Path is expected to be relative to `root` because we use StripPrefix("/static/")
		clean := path.Clean(strings.TrimPrefix(r.URL.Path, "/"))

		// We always vary on Accept-Encoding for correctness and cacheability
		w.Header().Add("Vary", "Accept-Encoding")

		ae := r.Header.Get("Accept-Encoding")
		supportsBR := strings.Contains(ae, "br")
		supportsGZ := strings.Contains(ae, "gzip")

		// Attempt Brotli first (best compression & widely supported by modern browsers)
		if supportsBR {
			if _, err := fs.Stat(root, clean+".br"); err == nil {
				// Set Content-Type based on the original (uncompressed) extension
				if ctype := mime.TypeByExtension(path.Ext(clean)); ctype != "" {
					w.Header().Set("Content-Type", ctype)
				}
				// Tell the client what they are getting
				w.Header().Set("Content-Encoding", "br")

				// Long-lived, immutable cache is safe if you fingerprint filenames (recommended)
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

				// Internally rewrite to the .br object and pass to the standard file server
				r2 := r.Clone(r.Context())
				r2.URL = &url.URL{Path: r.URL.Path + ".br", RawQuery: r.URL.RawQuery}
				base.ServeHTTP(w, r2)
				return
			}
		}

		// Fall back to gzip if available
		if supportsGZ {
			if _, err := fs.Stat(root, clean+".gz"); err == nil {
				if ctype := mime.TypeByExtension(path.Ext(clean)); ctype != "" {
					w.Header().Set("Content-Type", ctype)
				}
				w.Header().Set("Content-Encoding", "gzip")
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

				r2 := r.Clone(r.Context())
				r2.URL = &url.URL{Path: r.URL.Path + ".gz", RawQuery: r.URL.RawQuery}
				base.ServeHTTP(w, r2)
				return
			}
		}

		// No precompressed variant – serve the original asset
		// You may still want some sensible caching here if filenames are fingerprinted.
		// w.Header().Set("Cache-Control", "public, max-age=86400")
		base.ServeHTTP(w, r)
	})
}
