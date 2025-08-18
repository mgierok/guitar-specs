#!/bin/bash

# Simple frontend build script using esbuild and Tailwind CSS
# This script can be used when npm is not available

set -e

echo "→ Building frontend assets..."

# Create output directories
mkdir -p web/static/dist/js
mkdir -p web/static/dist/css

# Build JavaScript with esbuild (if available)
if command -v esbuild >/dev/null 2>&1; then
    echo "→ Building JavaScript with esbuild..."
    esbuild web/static/src/main.js \
        --bundle \
        --minify \
        --outdir=web/static/dist/js \
        --format=esm
else
    echo "⚠️  esbuild not found, copying source JavaScript..."
    cp web/static/src/main.js web/static/dist/js/main.js
fi

# Build CSS with Tailwind (if available)
if command -v tailwindcss >/dev/null 2>&1; then
    echo "→ Building CSS with Tailwind..."
    tailwindcss -i web/static/src/style.css -o web/static/dist/css/style.css --minify
else
    echo "⚠️  Tailwind CSS not found, copying source CSS..."
    cp web/static/src/style.css web/static/dist/css/style.css
fi

echo "✓ Frontend assets built successfully"
