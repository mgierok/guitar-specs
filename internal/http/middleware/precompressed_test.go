package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

// British English: helper that serves /static/* using the precompressed handler.
// It optionally wraps with a Cache-Control middleware to mimic the app's static group.
func testMux(root fstest.MapFS, withCache bool) http.Handler {
	h := http.StripPrefix("/static/", PrecompressedFileServer(root))
	if withCache {
		h = withCacheControl(h)
	}
	mux := http.NewServeMux()
	mux.Handle("/static/", h)
	return mux
}

// British English: sets an immutable Cache-Control header, as used for fingerprinted assets.
func withCacheControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
}

func TestPrecompressed_Brotli_SetsEncodingAndCache(t *testing.T) {
	fs := fstest.MapFS{
		"js/app.js":    &fstest.MapFile{Data: []byte("alert(1)")},
		"js/app.js.br": &fstest.MapFile{Data: []byte{0x1b, 0x2a}}, // British English: dummy payload is fine for the test
	}
	mux := testMux(fs, false) // Cache-Control for br/gz is set by the handler itself.

	req := httptest.NewRequest(http.MethodGet, "/static/js/app.js", nil)
	req.Header.Set("Accept-Encoding", "br")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if ce := rr.Header().Get("Content-Encoding"); ce != "br" {
		t.Fatalf("expected Content-Encoding=br, got %q", ce)
	}
	if vary := rr.Header().Get("Vary"); !strings.Contains(vary, "Accept-Encoding") {
		t.Fatalf("expected Vary to contain Accept-Encoding, got %q", vary)
	}
	if cc := rr.Header().Get("Cache-Control"); !strings.Contains(cc, "immutable") {
		t.Fatalf("expected Cache-Control immutable for br path, got %q", cc)
	}
}

func TestPrecompressed_Gzip_SetsEncodingAndCache(t *testing.T) {
	fs := fstest.MapFS{
		"js/app.js":    &fstest.MapFile{Data: []byte("alert(1)")},
		"js/app.js.gz": &fstest.MapFile{Data: []byte{0x1f, 0x8b}}, // British English: dummy gzip header for the test
	}
	mux := testMux(fs, false)

	req := httptest.NewRequest(http.MethodGet, "/static/js/app.js", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if ce := rr.Header().Get("Content-Encoding"); ce != "gzip" {
		t.Fatalf("expected Content-Encoding=gzip, got %q", ce)
	}
	if vary := rr.Header().Get("Vary"); !strings.Contains(vary, "Accept-Encoding") {
		t.Fatalf("expected Vary to contain Accept-Encoding, got %q", vary)
	}
	if cc := rr.Header().Get("Cache-Control"); !strings.Contains(cc, "immutable") {
		t.Fatalf("expected Cache-Control immutable for gzip path, got %q", cc)
	}
}

func TestPrecompressed_Original_NoEncoding_HeaderFromAppGroup(t *testing.T) {
	fs := fstest.MapFS{
		"js/app.js": &fstest.MapFile{Data: []byte("alert(1)")},
	}
	// British English: simulate the app's static group that sets Cache-Control for originals.
	mux := testMux(fs, true)

	req := httptest.NewRequest(http.MethodGet, "/static/js/app.js", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if ce := rr.Header().Get("Content-Encoding"); ce != "" {
		t.Fatalf("expected no Content-Encoding for original asset, got %q", ce)
	}
	if cc := rr.Header().Get("Cache-Control"); !strings.Contains(cc, "immutable") {
		t.Fatalf("expected Cache-Control immutable from app group, got %q", cc)
	}
}
