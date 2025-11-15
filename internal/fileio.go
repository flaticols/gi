package internal

import (
	"io"
	"os"
)

// FileIO provides an interface for file I/O operations that can be mocked for WASM.
type FileIO interface {
	ReadFile(filename string) ([]byte, error)
	WriteFile(filename string, data []byte, perm os.FileMode) error
}

// OSFileIO implements FileIO using the standard os package.
type OSFileIO struct{}

func (OSFileIO) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (OSFileIO) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

// MemFileIO implements FileIO using an in-memory file system for WASM.
type MemFileIO struct {
	files map[string][]byte
}

func NewMemFileIO() *MemFileIO {
	return &MemFileIO{
		files: make(map[string][]byte),
	}
}

func (m *MemFileIO) ReadFile(filename string) ([]byte, error) {
	data, ok := m.files[filename]
	if !ok {
		return nil, &os.PathError{Op: "open", Path: filename, Err: os.ErrNotExist}
	}
	return data, nil
}

func (m *MemFileIO) WriteFile(filename string, data []byte, perm os.FileMode) error {
	m.files[filename] = data
	return nil
}

// AddFile adds a file to the in-memory filesystem.
func (m *MemFileIO) AddFile(filename string, data []byte) {
	m.files[filename] = data
}

// ErrorHandler provides an interface for handling errors that can be customized for WASM.
type ErrorHandler interface {
	HandleFatalError(err error, frames []string)
}

// OSErrorHandler implements ErrorHandler using os.Exit.
type OSErrorHandler struct {
	Stderr io.Writer
}

func NewOSErrorHandler() *OSErrorHandler {
	return &OSErrorHandler{Stderr: os.Stderr}
}

func (h *OSErrorHandler) HandleFatalError(err error, frames []string) {
	if h.Stderr == nil {
		h.Stderr = os.Stderr
	}
	io.WriteString(h.Stderr, "[gi] fatal error: ")
	io.WriteString(h.Stderr, err.Error())
	io.WriteString(h.Stderr, "\n\n")
	for _, frame := range frames {
		io.WriteString(h.Stderr, "[gi] ")
		io.WriteString(h.Stderr, frame)
		io.WriteString(h.Stderr, "\n")
	}
	os.Exit(1)
}

// PanicErrorHandler implements ErrorHandler using panic instead of os.Exit for WASM.
type PanicErrorHandler struct {
	Stderr io.Writer
}

func NewPanicErrorHandler() *PanicErrorHandler {
	return &PanicErrorHandler{Stderr: os.Stderr}
}

func (h *PanicErrorHandler) HandleFatalError(err error, frames []string) {
	if h.Stderr == nil {
		h.Stderr = os.Stderr
	}
	io.WriteString(h.Stderr, "[gi] fatal error: ")
	io.WriteString(h.Stderr, err.Error())
	io.WriteString(h.Stderr, "\n\n")
	for _, frame := range frames {
		io.WriteString(h.Stderr, "[gi] ")
		io.WriteString(h.Stderr, frame)
		io.WriteString(h.Stderr, "\n")
	}
	panic(err)
}

// Global instances - can be replaced for WASM
var (
	defaultFileIO      FileIO        = OSFileIO{}
	defaultErrorHandler ErrorHandler = NewOSErrorHandler()
)

// SetFileIO sets the global FileIO implementation.
func SetFileIO(fio FileIO) {
	defaultFileIO = fio
}

// SetErrorHandler sets the global ErrorHandler implementation.
func SetErrorHandler(eh ErrorHandler) {
	defaultErrorHandler = eh
}

// GetFileIO returns the current FileIO implementation.
func GetFileIO() FileIO {
	return defaultFileIO
}

// GetErrorHandler returns the current ErrorHandler implementation.
func GetErrorHandler() ErrorHandler {
	return defaultErrorHandler
}
