# Precompress static assets (gzip + brotli)
assets-precompress:
	@echo "→ Precompressing /web/static assets..."
	@find web/static -type f \
	  \( -name '*.js' -o -name '*.css' -o -name '*.svg' -o -name '*.html' -o -name '*.json' -o -name '*.wasm' \) \
	  -print0 | xargs -0 -I{} sh -c 'gzip -k -f -9 "{}"; brotli -f -q 11 "{}"'
	@echo "✓ Done."

run:
	go generate ./...
	go run ./cmd/web

build:
	go generate ./...
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o bin/web ./cmd/web

test:
	go test ./...
