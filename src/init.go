package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const defaultConfigYAML = `gemini_command: "gemini"
verify_command: "python3 -m py_compile utils.py"
success_command: "echo 'Successfully implemented: $CANDIDATE'"
reset_command: "echo 'Reset (no changes persisted)'"
`

const defaultTaskYAML = `candidate_source: |
  # Find all TODO comments and output as JSON array
  grep -n "# TODO:" utils.py 2>/dev/null | while read line; do
    echo "$line"
  done | jq -R -s 'split("\n") | map(select(length > 0))'

prompt: |
  Implement the TODO in utils.py.

  The TODO to implement:
  $INPUT

  Instructions:
  1. Read utils.py to understand the context
  2. Implement the functionality described in the TODO comment
  3. Remove the TODO comment after implementing
  4. Make sure the code is syntactically correct

  Important: Only implement THIS specific TODO, not others.
`

const defaultUtilsPy = `"""Utility functions for data processing."""


def add(a, b):
    """Add two numbers."""
    return a + b


def subtract(a, b):
    # TODO: Implement subtraction
    pass


def multiply(a, b):
    # TODO: Implement multiplication
    pass


def divide(a, b):
    # TODO: Implement division with zero-check
    pass


def is_even(n):
    # TODO: Return True if n is even, False otherwise
    pass


def factorial(n):
    # TODO: Implement factorial (n! = n * (n-1) * ... * 1)
    pass
`

// InitializeAgena creates a starter agena/ layout in cwd.
// Existing files are preserved; only missing files are created.
func InitializeAgena(cwd string) ([]string, error) {
	runnerDir := filepath.Join(cwd, "agena")
	taskDir := filepath.Join(runnerDir, "fix-todos")

	if err := os.MkdirAll(taskDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create %s: %w", taskDir, err)
	}

	var created []string
	configPath := filepath.Join(runnerDir, "config.yaml")
	taskPath := filepath.Join(taskDir, "task.yaml")
	utilsPath := filepath.Join(cwd, "utils.py")

	if ok, err := writeFileIfMissing(configPath, defaultConfigYAML); err != nil {
		return nil, err
	} else if ok {
		created = append(created, "agena/config.yaml")
	}

	if ok, err := writeFileIfMissing(taskPath, defaultTaskYAML); err != nil {
		return nil, err
	} else if ok {
		created = append(created, "agena/fix-todos/task.yaml")
	}

	if ok, err := writeFileIfMissing(utilsPath, defaultUtilsPy); err != nil {
		return nil, err
	} else if ok {
		created = append(created, "utils.py")
	}

	return created, nil
}

func writeFileIfMissing(path, content string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("failed to check %s: %w", path, err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return false, fmt.Errorf("failed to write %s: %w", path, err)
	}
	return true, nil
}
