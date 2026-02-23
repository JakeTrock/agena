package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitializeAgenaCreatesStarterFiles(t *testing.T) {
	cwd := t.TempDir()

	created, err := InitializeAgena(cwd)
	if err != nil {
		t.Fatalf("InitializeAgena() error = %v", err)
	}
	if len(created) != 3 {
		t.Fatalf("expected 3 created files, got %d (%v)", len(created), created)
	}

	if _, err := os.Stat(filepath.Join(cwd, "agena", "config.yaml")); err != nil {
		t.Fatalf("expected agena/config.yaml to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(cwd, "agena", "fix-todos", "task.yaml")); err != nil {
		t.Fatalf("expected agena/fix-todos/task.yaml to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(cwd, "utils.py")); err != nil {
		t.Fatalf("expected utils.py to exist: %v", err)
	}
}

func TestInitializeAgenaDoesNotOverwriteExistingFiles(t *testing.T) {
	cwd := t.TempDir()
	configPath := filepath.Join(cwd, "agena", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatal(err)
	}

	original := "gemini_command: \"custom\"\n"
	if err := os.WriteFile(configPath, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := InitializeAgena(cwd)
	if err != nil {
		t.Fatalf("InitializeAgena() error = %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != original {
		t.Fatalf("expected existing config to be preserved, got: %q", string(data))
	}
}
