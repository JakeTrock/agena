package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const defaultConfigYAML = `gemini_command: "gemini"
verify_command: "echo 'set verify_command in agena/config.yaml'"
success_command: "git commit -m 'Fix: $CANDIDATE'"
reset_command: "git reset --hard"
`

const defaultTaskYAML = `candidate_source: "echo '[\"example-candidate\"]'"
prompt: |
  Replace this task with your real workflow.
  Candidate: $INPUT
`

// InitializeAgena creates a starter agena/ layout in cwd.
// Existing files are preserved; only missing files are created.
func InitializeAgena(cwd string) ([]string, error) {
	runnerDir := filepath.Join(cwd, "agena")
	taskDir := filepath.Join(runnerDir, "example-task")

	if err := os.MkdirAll(taskDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create %s: %w", taskDir, err)
	}

	var created []string
	configPath := filepath.Join(runnerDir, "config.yaml")
	taskPath := filepath.Join(taskDir, "task.yaml")

	if ok, err := writeFileIfMissing(configPath, defaultConfigYAML); err != nil {
		return nil, err
	} else if ok {
		created = append(created, "agena/config.yaml")
	}

	if ok, err := writeFileIfMissing(taskPath, defaultTaskYAML); err != nil {
		return nil, err
	} else if ok {
		created = append(created, "agena/example-task/task.yaml")
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
