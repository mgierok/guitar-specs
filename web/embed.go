//go:generate go run ../cmd/genstatic -src=static -brotli -gzip

package web

import (
	"embed"
)

//go:embed templates/**/*
var TemplatesFS embed.FS

//go:embed static
var StaticFS embed.FS

//go:embed robots.txt
var RobotsFS embed.FS
