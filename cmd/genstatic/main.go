package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	// British English: pure-Go Brotli encoder, no CGO needed.
	"github.com/andybalholm/brotli"
)

type options struct {
	src       string
	doBr      bool
	doGz      bool
	brQuality int
	gzLevel   int
}

func main() {
	var opt options
	flag.StringVar(&opt.src, "src", "web/static", "source directory with static assets")
	flag.BoolVar(&opt.doBr, "brotli", false, "generate .br alongside originals")
	flag.BoolVar(&opt.doGz, "gzip", false, "generate .gz alongside originals")
	flag.IntVar(&opt.brQuality, "brq", 11, "brotli quality (0-11)")
	flag.IntVar(&opt.gzLevel, "gzq", 9, "gzip level (1-9)")
	flag.Parse()

	if !opt.doBr && !opt.doGz {
		fmt.Fprintln(os.Stderr, "nothing to do: enable -brotli and/or -gzip")
		os.Exit(2)
	}

	var total, brCnt, gzCnt int
	err := filepath.Walk(opt.src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// British English: we only compress text-like assets; skip already-compressed or binary types.
		low := strings.ToLower(p)
		if strings.HasSuffix(low, ".br") || strings.HasSuffix(low, ".gz") {
			return nil
		}
		if !shouldCompress(low) {
			return nil
		}

		total++
		if opt.doBr {
			if err := maybeBrotli(p, opt.brQuality); err != nil {
				return fmt.Errorf("brotli %s: %w", p, err)
			}
			brCnt++
		}
		if opt.doGz {
			if err := maybeGzip(p, opt.gzLevel); err != nil {
				return fmt.Errorf("gzip %s: %w", p, err)
			}
			gzCnt++
		}
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "genstatic failed:", err)
		os.Exit(1)
	}

	fmt.Printf("genstatic: scanned=%d, br=%d, gz=%d\n", total, brCnt, gzCnt)
}

// British English: choose only text-like assets that benefit from compression.
func shouldCompress(path string) bool {
	switch {
	case strings.HasSuffix(path, ".js"),
		strings.HasSuffix(path, ".mjs"),
		strings.HasSuffix(path, ".css"),
		strings.HasSuffix(path, ".svg"),
		strings.HasSuffix(path, ".json"),
		strings.HasSuffix(path, ".wasm"),
		strings.HasSuffix(path, ".txt"),
		strings.HasSuffix(path, ".xml"):
		return true
	}
	return false
}

// British English: write .br if missing or older than the source.
func maybeBrotli(src string, quality int) error {
	dst := src + ".br"
	if upToDate(src, dst) {
		return nil
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	tmp := dst + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}

	w := brotli.NewWriterLevel(out, quality)
	if _, err := io.Copy(w, in); err != nil {
		_ = w.Close()
		_ = out.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := w.Close(); err != nil {
		_ = out.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}

	return os.Rename(tmp, dst)
}

// British English: write .gz if missing or older than the source.
func maybeGzip(src string, level int) error {
	dst := src + ".gz"
	if upToDate(src, dst) {
		return nil
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	tmp := dst + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}

	w, err := gzip.NewWriterLevel(out, level)
	if err != nil {
		_ = out.Close()
		_ = os.Remove(tmp)
		return err
	}
	// Set original name for better tooling compatibility (optional).
	w.Name = filepath.Base(src)

	if _, err := io.Copy(w, in); err != nil {
		_ = w.Close()
		_ = out.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := w.Close(); err != nil {
		_ = out.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}

	return os.Rename(tmp, dst)
}

// British English: true when destination exists and is newer than source.
func upToDate(src, dst string) bool {
	si, err1 := os.Stat(src)
	di, err2 := os.Stat(dst)
	if err1 != nil || err2 != nil {
		return false
	}
	return !di.ModTime().Before(si.ModTime().Add(-1 * time.Second))
}
