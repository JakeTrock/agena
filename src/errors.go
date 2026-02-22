package main

import (
	"errors"
	"fmt"
	"strings"
)

// formatUserError converts internal errors into actionable CLI guidance.
func formatUserError(err error) string {
	if err == nil {
		return "unknown error"
	}

	msg := err.Error()

	if strings.Contains(msg, "task not found:") {
		return fmt.Sprintf("%s\nUse `agena --list` to see available tasks in this project.", msg)
	}

	if strings.Contains(msg, "failed to load config: failed to read") && strings.Contains(msg, "config.yaml") {
		return strings.Join([]string{
			"Couldn't read your Agena config file.",
			msg,
			"",
			"How to fix:",
			"  1. Run this command from your project root (the folder that should contain `agena/`).",
			"  2. If this is a new project, run `agena --init` to create starter files.",
			"  3. Ensure `agena/config.yaml` exists and is readable.",
		}, "\n")
	}

	if strings.Contains(msg, "failed to initialize agena directory:") && strings.Contains(msg, "not a directory") {
		return strings.Join([]string{
			"Agena couldn't create its `agena/` setup in this folder because something named `agena` already exists as a file.",
			"How to fix:",
			"  1. Run Agena from your project root.",
			"  2. Keep the binary on PATH and run `agena` from the repository you want to work on.",
			"",
			msg,
		}, "\n")
	}

	if strings.Contains(msg, "not a directory") && strings.Contains(msg, "config.yaml") {
		return strings.Join([]string{
			"Agena expected an `agena/` directory, but found a file named `agena` instead.",
			msg,
			"",
			"How to fix:",
			"  1. Run Agena from your project root, not from the `bin/` folder.",
			"  2. If needed, create setup with `agena --init`.",
		}, "\n")
	}

	if strings.Contains(msg, "gemini command not found:") {
		return fmt.Sprintf("%s\nInstall Gemini CLI and ensure it is in PATH, or set `gemini_command` in `agena/config.yaml`.", msg)
	}

	var interpErr *interpolationError
	if errors.As(err, &interpErr) {
		return strings.Join([]string{
			msg,
			"Update your task prompt so `$INPUT` usage matches the candidate data shape.",
			"For example, use `$INPUT` for strings and `$INPUT[0]` only for array candidates.",
		}, "\n")
	}

	return msg
}
