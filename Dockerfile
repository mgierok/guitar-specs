# Build stage
FROM golang:1.25 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Install brotli for precompression (gzip is available by default)
RUN apt-get update && apt-get install -y --no-install-recommends brotli && rm -rf /var/lib/apt/lists/*

# Build the app
RUN go generate ./... && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/web ./cmd/web

# Run stage
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/web /app/web
EXPOSE 8443
USER nonroot:nonroot
ENTRYPOINT ["/app/web"]
