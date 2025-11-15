# WebAssembly (WASM) Support

This document explains how to compile and use the `gi` Go interpreter with WebAssembly.

## Overview

The `gi` interpreter can be compiled to WebAssembly, allowing you to run Go code directly in web browsers. This is useful for creating interactive Go playgrounds, educational tools, and browser-based development environments.

## Key Considerations for WASM

When running in a WASM environment, several platform-specific operations need special handling:

### 1. File I/O Operations

WASM runs in a sandboxed environment without direct file system access. The `gi` library provides an abstraction layer for file I/O:

- **`FileIO` interface**: Abstracts file read/write operations
- **`OSFileIO`**: Default implementation using `os.ReadFile`/`os.WriteFile` (for native platforms)
- **`MemFileIO`**: In-memory implementation for WASM environments

### 2. Process Control

The `os.Exit()` function doesn't work the same way in WASM. Instead:

- **`ErrorHandler` interface**: Abstracts fatal error handling
- **`OSErrorHandler`**: Uses `os.Exit(1)` for native platforms
- **`PanicErrorHandler`**: Uses `panic()` for WASM, allowing JavaScript to catch errors

### 3. Setting Up for WASM

Before running the interpreter in WASM, configure the appropriate handlers:

```go
import (
    "github.com/emicklei/gi/internal"
)

func main() {
    // Use panic instead of os.Exit for WASM
    internal.SetErrorHandler(internal.NewPanicErrorHandler())
    
    // Use in-memory file system
    memFS := internal.NewMemFileIO()
    internal.SetFileIO(memFS)
    
    // Your WASM code here...
}
```

## Building for WASM

### Prerequisites

- Go 1.25 or later
- A modern web browser with WebAssembly support

### Build Steps

1. **Set the target environment variables:**
   ```bash
   GOOS=js GOARCH=wasm go build -o main.wasm main.go
   ```

2. **Copy the WASM JavaScript support file:**
   ```bash
   cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
   ```

3. **Serve the files:**
   ```bash
   # Using Python
   python3 -m http.server 8080
   
   # Or using Node.js
   npx serve
   ```

### Example Application

See the complete example in `examples/wasm/`:

```bash
cd examples/wasm
./build.sh
python3 -m http.server 8080
# Open http://localhost:8080 in your browser
```

## Using the Abstraction Layer

### Custom File I/O

If you need to provide files to the interpreter in WASM:

```go
memFS := internal.NewMemFileIO()

// Add files to the in-memory filesystem
memFS.AddFile("config.json", []byte(`{"key": "value"}`))
memFS.AddFile("data.txt", []byte("Hello, World!"))

internal.SetFileIO(memFS)
```

### Custom Error Handling

For custom error handling behavior:

```go
type CustomErrorHandler struct {
    OnError func(error, []string)
}

func (h *CustomErrorHandler) HandleFatalError(err error, frames []string) {
    if h.OnError != nil {
        h.OnError(err, frames)
    }
    // Handle error as needed for your environment
}

internal.SetErrorHandler(&CustomErrorHandler{
    OnError: func(err error, frames []string) {
        // Log to JavaScript console, send to analytics, etc.
    },
})
```

## Limitations in WASM

When running in WASM, be aware of these limitations:

1. **No direct file system access**: Use `MemFileIO` or implement a virtual file system
2. **No `os.Exit()`**: Errors will panic instead, which can be caught by JavaScript
3. **Package loading**: The `golang.org/x/tools/go/packages` loader may not work in WASM for loading external packages from disk
4. **Limited stdlib access**: Some Go standard library functions may not work in WASM (e.g., those requiring OS-specific syscalls)

## Architecture

The WASM support is implemented through these key abstractions:

```
┌─────────────────────────────────────────┐
│         gi Application Code              │
└─────────────────┬───────────────────────┘
                  │
         ┌────────▼─────────┐
         │  Abstraction     │
         │    Interfaces    │
         └────────┬─────────┘
                  │
     ┌────────────┴────────────┐
     │                         │
┌────▼────────┐      ┌────────▼────────┐
│   Native    │      │      WASM       │
│ (OSFileIO,  │      │  (MemFileIO,    │
│ OSErrorHandler)    │ PanicErrorHandler)│
└─────────────┘      └─────────────────┘
```

## Testing

To test WASM functionality without a browser, you can use the `wasmer` or `wasmtime` runtimes:

```bash
# Install wasmer
curl https://get.wasmer.io -sSfL | sh

# Run the WASM binary
wasmer main.wasm
```

## Best Practices

1. **Always set handlers early**: Configure `FileIO` and `ErrorHandler` at the start of your WASM `main()` function
2. **Use defer/recover**: Wrap interpreter calls in `defer/recover` to gracefully handle panics
3. **Provide virtual files**: Pre-populate the `MemFileIO` with any files your Go code needs
4. **Test in target browsers**: Different browsers may have varying WASM support levels

## Future Enhancements

Potential improvements for better WASM support:

- Virtual package loader for importing standard library packages
- Persistent storage using IndexedDB or localStorage
- WebWorker support for concurrent execution
- Streaming compilation for large Go programs
