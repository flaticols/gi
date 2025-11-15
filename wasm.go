package gi

import "github.com/emicklei/gi/internal"

// SetFileIO configures the file I/O implementation used by the interpreter.
// Use this to provide a custom file system implementation for WASM or other
// environments where standard file I/O is not available.
//
// Example for WASM:
//
//	memFS := internal.NewMemFileIO()
//	memFS.AddFile("config.json", []byte(`{"key": "value"}`))
//	gi.SetFileIO(memFS)
func SetFileIO(fio internal.FileIO) {
	internal.SetFileIO(fio)
}

// SetErrorHandler configures how fatal errors are handled by the interpreter.
// Use this to customize error handling for different environments.
//
// For WASM, use PanicErrorHandler instead of the default OSErrorHandler:
//
//	gi.SetErrorHandler(internal.NewPanicErrorHandler())
//
// This allows JavaScript to catch errors instead of calling os.Exit().
func SetErrorHandler(eh internal.ErrorHandler) {
	internal.SetErrorHandler(eh)
}

// NewMemFileIO creates a new in-memory file system suitable for WASM environments.
// Files can be added using the AddFile method before running the interpreter.
//
// Example:
//
//	memFS := gi.NewMemFileIO()
//	memFS.AddFile("data.txt", []byte("Hello, World!"))
//	gi.SetFileIO(memFS)
func NewMemFileIO() *internal.MemFileIO {
	return internal.NewMemFileIO()
}

// NewPanicErrorHandler creates an error handler that uses panic instead of os.Exit.
// This is suitable for WASM environments where os.Exit is not appropriate.
//
// Example:
//
//	gi.SetErrorHandler(gi.NewPanicErrorHandler())
func NewPanicErrorHandler() *internal.PanicErrorHandler {
	return internal.NewPanicErrorHandler()
}
