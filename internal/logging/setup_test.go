package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetupWritesToFile(t *testing.T) {
	dir := t.TempDir()

	log, closer, err := Setup(Options{
		Level:   "info",
		Env:     "development",
		LogDir:  dir,
		LogFile: "test.log",
	})
	if err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	defer closer.Close()

	log.Info("file log test", "component", "logging")

	path := filepath.Join(dir, "test.log")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "file log test") {
		t.Errorf("log file missing message: %q", content)
	}
}

func TestSetupAppendsToExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "append.log")
	if err := os.WriteFile(path, []byte("existing\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	log, closer, err := Setup(Options{
		Level:   "info",
		Env:     "development",
		LogDir:  dir,
		LogFile: "append.log",
	})
	if err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	defer closer.Close()

	log.Info("appended line")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "existing") {
		t.Error("expected existing content to remain")
	}
	if !strings.Contains(content, "appended line") {
		t.Error("expected appended content")
	}
}
