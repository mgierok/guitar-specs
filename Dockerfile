# Build stage
FROM golang:1.25 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build the app (no precompression step)
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/web ./cmd/web

# Run stage
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/web /app/web
EXPOSE 8443
USER nonroot:nonroot
ENTRYPOINT ["/app/web"]
