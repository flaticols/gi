# gi WASM Example

This example demonstrates how to compile and use the `gi` Go interpreter in a web browser using WebAssembly.

## Quick Start

1. **Build the WASM binary:**
   ```bash
   ./build.sh
   ```

2. **Start a local web server:**
   ```bash
   python3 -m http.server 8080
   # or
   npx serve
   ```

3. **Open in your browser:**
   Navigate to http://localhost:8080

## What's Included

- `main.go` - WASM entry point that exposes the `gi` interpreter to JavaScript
- `index.html` - Web interface for writing and running Go code
- `build.sh` - Build script for compiling to WASM
- `go.mod` - Go module file with local replacement for `gi`

## How It Works

1. The Go code is compiled to WebAssembly using `GOOS=js GOARCH=wasm`
2. The WASM module is loaded in the browser using Go's `wasm_exec.js`
3. A JavaScript function `executeGo(sourceCode)` is exposed to run Go code
4. The web interface provides an editor and output console

## Features

- Write Go code directly in the browser
- Instant execution without server round-trips
- Capture and display console output
- Error handling and display

## Architecture

```
Browser
  │
  ├─ HTML/JS Interface
  │     │
  │     └─ calls executeGo(source)
  │
  └─ WASM Module (main.wasm)
        │
        ├─ gi.ParseSource(source)
        ├─ gi.Call(pkg, "main")
        └─ Uses MemFileIO + PanicErrorHandler
```

## Customization

### Adding Virtual Files

To make files available to the interpreted Go code:

```go
memFS := gi.NewMemFileIO()
memFS.AddFile("config.json", []byte(`{"setting": "value"}`))
gi.SetFileIO(memFS)
```

### Custom Error Handling

```go
customHandler := gi.NewPanicErrorHandler()
customHandler.Stderr = customWriter // your custom writer
gi.SetErrorHandler(customHandler)
```

## Troubleshooting

### WASM fails to load
- Ensure your web server serves `.wasm` files with the correct MIME type: `application/wasm`
- Check browser console for detailed error messages

### Code doesn't execute
- Verify the WASM module has loaded (check the status message)
- Check that your Go code has a `package main` and `func main()`

### Import errors
- The WASM environment has limited access to standard library packages
- Some packages that rely on OS features may not work

## Performance Notes

- First execution may be slower due to WASM compilation
- Subsequent executions are faster as the module is cached
- Complex Go code may take longer to parse and execute

## Browser Compatibility

Tested with:
- Chrome/Chromium 90+
- Firefox 89+
- Safari 14+
- Edge 90+

## Learn More

See the main [WASM documentation](../../docs/WASM.md) for detailed information about WASM support in `gi`.
