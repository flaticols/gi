package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/emicklei/gi/internal"
)

//go:embed static/*
var staticFiles embed.FS

var (
	addr    = flag.String("addr", ":8080", "HTTP server address")
	timeout = flag.Duration("timeout", 5*time.Second, "execution timeout")
)

func main() {
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/run", handleRun)
	mux.HandleFunc("/share", handleShare)
	mux.HandleFunc("/p/", handleLoad)
	mux.HandleFunc("/embed", handleEmbed)
	mux.Handle("/static/", http.FileServer(http.FS(staticFiles)))

	log.Printf("Starting gi-playground on %s", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	serveHTML(w, r, false)
}

func handleEmbed(w http.ResponseWriter, r *http.Request) {
	serveHTML(w, r, true)
}

func serveHTML(w http.ResponseWriter, r *http.Request, isEmbed bool) {
	// Get initial code from query parameter if present
	code := r.URL.Query().Get("code")
	if code == "" {
		code = defaultCode
	} else {
		// Decode base64 code
		decoded, err := base64.URLEncoding.DecodeString(code)
		if err == nil {
			code = string(decoded)
		}
	}

	tmpl := indexHTML
	if isEmbed {
		tmpl = embedHTML
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Replace placeholder with actual code
	html := strings.ReplaceAll(tmpl, "{{CODE}}", escapeJS(code))
	fmt.Fprint(w, html)
}

type runRequest struct {
	Code string `json:"code"`
}

type runResponse struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

func handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req runRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, runResponse{Error: "Invalid request body"})
		return
	}

	if req.Code == "" {
		respondJSON(w, http.StatusBadRequest, runResponse{Error: "No code provided"})
		return
	}

	// Execute the code with timeout
	ctx, cancel := context.WithTimeout(r.Context(), *timeout)
	defer cancel()

	output, err := executeCode(ctx, req.Code)
	if err != nil {
		respondJSON(w, http.StatusOK, runResponse{Error: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, runResponse{Output: output})
}

type shareRequest struct {
	Code string `json:"code"`
}

type shareResponse struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

func handleShare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req shareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create a simple shareable ID by base64 encoding the code
	id := base64.URLEncoding.EncodeToString([]byte(req.Code))

	// Build the URL
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	host := r.Host

	resp := shareResponse{
		ID:  id,
		URL: fmt.Sprintf("%s://%s/p/%s", scheme, host, id),
	}

	respondJSON(w, http.StatusOK, resp)
}

func handleLoad(w http.ResponseWriter, r *http.Request) {
	// Extract the ID from the path
	id := strings.TrimPrefix(r.URL.Path, "/p/")
	if id == "" {
		http.Error(w, "No snippet ID provided", http.StatusBadRequest)
		return
	}

	// Decode the code
	code, err := base64.URLEncoding.DecodeString(id)
	if err != nil {
		http.Error(w, "Invalid snippet ID", http.StatusBadRequest)
		return
	}

	// Redirect to the main page with the code
	encoded := base64.URLEncoding.EncodeToString(code)
	http.Redirect(w, r, "/?code="+encoded, http.StatusFound)
}

func executeCode(ctx context.Context, code string) (string, error) {
	// Capture stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	os.Stdout = wOut
	os.Stderr = wErr

	// Channel for execution result
	done := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("panic: %v", r)
			}
		}()

		pkg, err := internal.ParseSource(code)
		if err != nil {
			done <- fmt.Errorf("parse error: %v", err)
			return
		}

		err = internal.RunPackageFunction(pkg, "main", nil)
		if err != nil {
			done <- fmt.Errorf("runtime error: %v", err)
			return
		}
		done <- nil
	}()

	var execErr error
	select {
	case <-ctx.Done():
		execErr = fmt.Errorf("execution timed out after %v", *timeout)
	case err := <-done:
		execErr = err
	}

	// Restore stdout and stderr
	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Read captured output
	var outBuf, errBuf bytes.Buffer
	io.Copy(&outBuf, rOut)
	io.Copy(&errBuf, rErr)
	rOut.Close()
	rErr.Close()

	output := outBuf.String()
	if errBuf.Len() > 0 {
		output += errBuf.String()
	}

	return output, execErr
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func escapeJS(s string) string {
	// Escape special characters for JavaScript string
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "`", "\\`")
	s = strings.ReplaceAll(s, "$", "\\$")
	return s
}

const defaultCode = `package main

import "fmt"

func main() {
	fmt.Println("Hello, Gi Playground!")
}
`

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Gi Playground</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <header>
        <h1>Gi Playground</h1>
        <p>The Go Interpreter Playground</p>
    </header>
    <main>
        <div class="toolbar">
            <button id="run-btn" onclick="runCode()">▶ Run</button>
            <button id="share-btn" onclick="shareCode()">📤 Share</button>
            <button id="format-btn" onclick="formatCode()">📋 Format</button>
            <span id="status"></span>
        </div>
        <div class="editor-container">
            <div class="editor-wrapper">
                <textarea id="code" spellcheck="false">` + "`{{CODE}}`" + `</textarea>
            </div>
            <div class="output-wrapper">
                <div class="output-header">Output</div>
                <pre id="output"></pre>
            </div>
        </div>
    </main>
    <footer>
        <p>Powered by <a href="https://github.com/emicklei/gi" target="_blank">Gi</a> - The Go Interpreter</p>
    </footer>
    <script src="/static/playground.js"></script>
</body>
</html>
`

const embedHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Gi Playground</title>
    <link rel="stylesheet" href="/static/embed.css">
</head>
<body class="embed">
    <div class="embed-container">
        <div class="toolbar">
            <span class="title">Gi Playground</span>
            <button id="run-btn" onclick="runCode()">▶ Run</button>
            <a href="/" target="_blank" class="open-link">Open in Playground</a>
        </div>
        <div class="editor-container">
            <textarea id="code" spellcheck="false">` + "`{{CODE}}`" + `</textarea>
            <div class="output-wrapper">
                <pre id="output"></pre>
            </div>
        </div>
    </div>
    <script src="/static/playground.js"></script>
</body>
</html>
`
