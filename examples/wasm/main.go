//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"syscall/js"

	"github.com/emicklei/gi"
)

func main() {
	// Set up WASM-compatible I/O handlers using the public API
	gi.SetErrorHandler(gi.NewPanicErrorHandler())
	memFS := gi.NewMemFileIO()
	gi.SetFileIO(memFS)

	// Create a channel to keep the Go program running
	done := make(chan struct{})

	// Register a function to execute Go code from JavaScript
	js.Global().Set("executeGo", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 1 {
			return map[string]interface{}{
				"error": "Expected exactly one argument (Go source code)",
			}
		}

		source := args[0].String()

		// Parse and execute the Go source
		pkg, err := gi.ParseSource(source)
		if err != nil {
			return map[string]interface{}{
				"error": fmt.Sprintf("Parse error: %v", err),
			}
		}

		// Use a panic handler to catch runtime errors instead of os.Exit
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Runtime error: %v\n", r)
			}
		}()

		// Execute the main function
		_, err = gi.Call(pkg, "main")
		if err != nil {
			return map[string]interface{}{
				"error": fmt.Sprintf("Execution error: %v", err),
			}
		}

		return map[string]interface{}{
			"success": true,
		}
	}))

	fmt.Println("Go WASM initialized. Use executeGo(sourceCode) from JavaScript.")

	// Keep the program running
	<-done
}
