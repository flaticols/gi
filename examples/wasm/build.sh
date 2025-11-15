#!/bin/bash

# Build the WASM binary
echo "Building WASM binary..."
GOOS=js GOARCH=wasm go build -o main.wasm main.go

# Copy wasm_exec.js from Go installation
echo "Copying wasm_exec.js..."
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .

echo "Build complete!"
echo ""
echo "To run the example:"
echo "1. Start a local web server in this directory:"
echo "   python3 -m http.server 8080"
echo "   or"
echo "   npx serve"
echo ""
echo "2. Open http://localhost:8080 in your browser"
