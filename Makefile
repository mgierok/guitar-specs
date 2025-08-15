run:
	go run ./cmd/web

build:
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o bin/web ./cmd/web

test:
	go test ./...

