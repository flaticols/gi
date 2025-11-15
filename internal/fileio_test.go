package internal

import (
	"os"
	"testing"
)

func TestMemFileIO(t *testing.T) {
	memFS := NewMemFileIO()

	// Test writing and reading
	testData := []byte("test content")
	err := memFS.WriteFile("test.txt", testData, 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	data, err := memFS.ReadFile("test.txt")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(data) != string(testData) {
		t.Errorf("Expected %q, got %q", testData, data)
	}

	// Test reading non-existent file
	_, err = memFS.ReadFile("nonexistent.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
	if !os.IsNotExist(err) {
		t.Errorf("Expected os.IsNotExist error, got: %v", err)
	}
}

func TestAddFile(t *testing.T) {
	memFS := NewMemFileIO()

	testData := []byte("added file content")
	memFS.AddFile("added.txt", testData)

	data, err := memFS.ReadFile("added.txt")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(data) != string(testData) {
		t.Errorf("Expected %q, got %q", testData, data)
	}
}

func TestOSFileIO(t *testing.T) {
	osFS := OSFileIO{}

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "osfileio-test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Test writing
	testData := []byte("test content")
	err = osFS.WriteFile(tmpFile.Name(), testData, 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Test reading
	data, err := osFS.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(data) != string(testData) {
		t.Errorf("Expected %q, got %q", testData, data)
	}
}

func TestSetGetFileIO(t *testing.T) {
	// Save original
	original := GetFileIO()
	defer SetFileIO(original)

	// Test setting and getting
	memFS := NewMemFileIO()
	SetFileIO(memFS)

	if GetFileIO() != memFS {
		t.Error("GetFileIO did not return the set FileIO")
	}
}

func TestSetGetErrorHandler(t *testing.T) {
	// Save original
	original := GetErrorHandler()
	defer SetErrorHandler(original)

	// Test setting and getting
	panicHandler := NewPanicErrorHandler()
	SetErrorHandler(panicHandler)

	if GetErrorHandler() != panicHandler {
		t.Error("GetErrorHandler did not return the set ErrorHandler")
	}
}

func TestPanicErrorHandler(t *testing.T) {
	handler := NewPanicErrorHandler()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic, but didn't panic")
		}
	}()

	handler.HandleFatalError(os.ErrInvalid, []string{"frame1", "frame2"})
}
